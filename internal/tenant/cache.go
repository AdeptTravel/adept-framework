package tenant

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/singleflight"

	"github.com/AdeptTravel/adept-framework/internal/metrics"
)

// Static defaults.  Override via env vars or a config package if desired.
const (
	IdleTTL       = 30 * time.Minute
	MaxEntries    = 100
	EvictInterval = 5 * time.Minute
)

// ErrNotFound is returned when a host is not present in the site table.
var ErrNotFound = errors.New("tenant not found")

// Cache lazily loads tenants, stores them in a sync.Map, and evicts them on
// idle TTL or LRU pressure.
type Cache struct {
	globalDB    *sqlx.DB
	sfg         singleflight.Group
	m           sync.Map
	evictTicker *time.Ticker
	idleTTL     time.Duration
	maxEntries  int
}

// New constructs a Cache and starts the background evictor.
func New(global *sqlx.DB, idleTTL time.Duration, maxEntries int) *Cache {
	c := &Cache{
		globalDB:   global,
		idleTTL:    idleTTL,
		maxEntries: maxEntries,
	}
	c.evictTicker = time.NewTicker(EvictInterval)
	go c.evictLoop()
	return c
}

// Get returns the Tenant for host, loading it on demand.
func (c *Cache) Get(host string) (*Tenant, error) {
	if v, ok := c.m.Load(host); ok {
		ent := v.(*entry)
		atomic.StoreInt64(&ent.lastSeen, time.Now().UnixNano())
		return ent.tenant, nil
	}

	v, err, _ := c.sfg.Do(host, func() (interface{}, error) {
		// Double-check after singleflight barrier.
		if v, ok := c.m.Load(host); ok {
			ent := v.(*entry)
			atomic.StoreInt64(&ent.lastSeen, time.Now().UnixNano())
			return ent.tenant, nil
		}
		ten, err := loadSite(context.Background(), c.globalDB, host)
		if err != nil {
			metrics.TenantLoadErrorsTotal.Inc()
			return nil, err
		}
		ent := &entry{
			tenant:   ten,
			lastSeen: time.Now().UnixNano(),
		}
		c.m.Store(host, ent)
		metrics.TenantLoadTotal.Inc()
		metrics.ActiveTenants.Inc()
		return ten, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*Tenant), nil
}
