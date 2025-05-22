// Cache implements a concurrency-safe, lazy-loading map of runtime tenants.
// Each tenant is loaded from the site table the first time its Host header
// appears, wrapped with a small connection pool, and stored in a sync.Map.
// A background evictor goroutine (see evictor.go) periodically removes
// idle tenants and trims the map to MaxEntries via LRU.
//
// This file adds comprehensive logging.  Every major lifecycle event
// (loading, not-found, load error, online, idle evict, LRU evict) is written
// through the *log.Logger provided to New.  When the server is run in a TTY
// those messages are also printed to stdout.
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

	"github.com/AdeptTravel/adept-framework/internal/metrics"
)

// --------------------------------------------------------------------
// Tunables
// --------------------------------------------------------------------

const (
	IdleTTL       = 30 * time.Minute // evict tenant after this idle duration
	MaxEntries    = 100              // cap cache; 0 disables size eviction
	EvictInterval = 5 * time.Minute  // evictor scan cadence
)

// --------------------------------------------------------------------
// Errors
// --------------------------------------------------------------------

var ErrNotFound = errors.New("tenant not found")

// --------------------------------------------------------------------
// Cache definition
// --------------------------------------------------------------------

type Cache struct {
	globalDB    *sqlx.DB
	log         *log.Logger
	sfg         singleflight.Group // coalesces concurrent loads per host
	m           sync.Map           // host → *entry
	evictTicker *time.Ticker
	idleTTL     time.Duration
	maxEntries  int
}

// New builds a Cache and starts its background evictor.  Pass a logger that
// writes to both the daily file and stdout (when interactive).
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

// Get looks up host in the cache, loading it on demand.  The call is entirely
// thread-safe and updates the entry’s last-seen timestamp each hit.
func (c *Cache) Get(host string) (*Tenant, error) {
	// Fast-path: present in map.
	if v, ok := c.m.Load(host); ok {
		ent := v.(*entry)
		atomic.StoreInt64(&ent.lastSeen, time.Now().UnixNano())
		return ent.tenant, nil
	}

	// Slow-path: singleflight load so only one goroutine hits the DB.
	v, err, _ := c.sfg.Do(host, func() (interface{}, error) {
		// Double-check after barrier to avoid duplicate load.
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
