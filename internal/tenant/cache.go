// internal/tenant/cache.go
//
// Tenant LRU cache.
//
// Context
// -------
// A Tenant represents one live site: its own DB pool, config map,
// Component router, and image variant cache.  To keep boot time near zero
// we *lazy-load* each tenant the first time its Host header appears.  A
// background goroutine evicts idle entries after `IdleTTL` or when the
// cache exceeds `MaxEntries`, using last-seen timestamps and LRU order.
//
// Workflow
// --------
//  1. `New` receives the control-plane DB, tunables, and a logger.
//     It starts one `evictLoop` goroutine that wakes every `EvictInterval`.
//  2. `Get(host)` checks the `sync.Map`.  If present it updates `lastSeen`
//     and returns the tenant pointer immediately.
//  3. On a miss it calls `singleflight.Group.Do`, so only one goroutine
//     performs the SQL load for a given host.  The loader uses
//     `loadSite` (see loader.go) to fetch the `meta.Record`, open the
//     tenant DB pool, and build the router.
//  4. Successful loads increment Prometheus counters; failures log and
//     bubble up.  Not-found hosts return `ErrNotFound`.
//  5. `evictLoop` scans the map:
//     • Remove entries idle > `idleTTL`.
//     • Trim LRU order to `maxEntries` if that cap is non-zero.
//
// Notes
// -----
// • All public methods are safe for concurrent use.
// • Oxford commas, two spaces after periods, no m-dash per Adept style.
package tenant

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/singleflight"

	"github.com/yanizio/adept/internal/metrics"
)

//
// Tunables
//

const (
	IdleTTL       = 30 * time.Minute // Evict tenant after this idle duration
	MaxEntries    = 100              // Cap cache; 0 disables size eviction
	EvictInterval = 5 * time.Minute  // Evictor scan cadence
)

//
// Errors
//

var ErrNotFound = errors.New("tenant not found")

//
// Cache definition
//

type Cache struct {
	globalDB    *sqlx.DB
	log         *log.Logger
	sfg         singleflight.Group // Coalesces concurrent loads per host
	m           sync.Map           // host → *entry
	evictTicker *time.Ticker
	idleTTL     time.Duration
	maxEntries  int
}

// New builds a Cache and starts its background evictor.
func New(global *sqlx.DB, idleTTL time.Duration, maxEntries int, lg *log.Logger) *Cache {
	c := &Cache{
		globalDB:   global,
		idleTTL:    idleTTL,
		maxEntries: maxEntries,
		log:        lg,
	}
	c.evictTicker = time.NewTicker(EvictInterval)
	go c.evictLoop()
	return c
}

// Get looks up host in the cache, loading it on demand.
func (c *Cache) Get(host string) (*Tenant, error) {
	// Fast-path: present in map.
	if v, ok := c.m.Load(host); ok {
		ent := v.(*entry)
		atomic.StoreInt64(&ent.lastSeen, time.Now().UnixNano())
		return ent.tenant, nil
	}

	// Slow-path: singleflight so only one goroutine hits the DB.
	v, err, _ := c.sfg.Do(host, func() (interface{}, error) {
		// Double-check after barrier.
		if v, ok := c.m.Load(host); ok {
			ent := v.(*entry)
			atomic.StoreInt64(&ent.lastSeen, time.Now().UnixNano())
			return ent.tenant, nil
		}

		c.log.Printf("tenant %s loading …", host)

		ten, err := loadSite(context.Background(), c.globalDB, host)
		if err == ErrNotFound {
			c.log.Printf("tenant %s not found in site table", host)
			metrics.TenantLoadErrorsTotal.Inc()
			return nil, err
		}
		if err != nil {
			c.log.Printf("tenant %s load error: %v", host, err)
			metrics.TenantLoadErrorsTotal.Inc()
			return nil, err
		}

		ent := &entry{
			tenant:   ten,
			lastSeen: time.Now().UnixNano(),
		}
		c.m.Store(host, ent)

		c.log.Printf("tenant %s online (pool opened)", host)
		metrics.TenantLoadTotal.Inc()
		metrics.ActiveTenants.Inc()
		return ten, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*Tenant), nil
}
