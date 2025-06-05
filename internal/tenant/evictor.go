// internal/tenant/evictor.go
//
// Background eviction loop for Tenant cache.
//
// Context
// -------
// The Tenant cache uses an `entry` wrapper that records `lastSeen`.  A
// single goroutine—started by `Cache.New()`—wakes every `EvictInterval`
// and prunes the map in two passes:
//
//   - **Idle eviction** — remove tenants idle longer than `idleTTL`.
//   - **LRU eviction**  — if the map still exceeds `maxEntries`, remove the
//     oldest entries until the cap is met.
//
// Each eviction closes the tenant’s DB pool, logs the event, and updates
// Prometheus metrics.
//
// Notes
// -----
//   - All map operations use `sync.Map` APIs; no additional locks needed.
//   - `metrics.TenantEvictTotal` and `metrics.ActiveTenants` gauge are
//     updated in-line.  This file imports only `time` and `sort` plus the
//     metrics package.
//   - Oxford commas, two spaces after periods, no m-dash.
package tenant

import (
	"sort"
	"sync/atomic"
	"time"

	"github.com/yanizio/adept/internal/metrics"
)

func (c *Cache) evictLoop() {
	for range c.evictTicker.C {
		now := time.Now().UnixNano()
		var count int

		//
		// Idle eviction pass
		//
		c.m.Range(func(key, value any) bool {
			count++
			ent := value.(*entry)
			idle := time.Duration(now-atomic.LoadInt64(&ent.lastSeen)) * time.Nanosecond
			if idle > c.idleTTL {
				_ = ent.tenant.Close()
				c.m.Delete(key)
				c.log.Printf("tenant %s evicted after %v idle", key, idle.Truncate(time.Second))
				metrics.TenantEvictTotal.Inc()
				metrics.ActiveTenants.Dec()
			}
			return true
		})

		//
		// LRU eviction pass
		//
		if c.maxEntries > 0 && count > c.maxEntries {
			type kv struct {
				key string
				at  int64
			}
			var all []kv
			c.m.Range(func(key, value any) bool {
				ent := value.(*entry)
				all = append(all, kv{key: key.(string), at: ent.lastSeen})
				return true
			})
			sort.Slice(all, func(i, j int) bool { return all[i].at < all[j].at })

			for i := 0; i < count-c.maxEntries; i++ {
				if v, ok := c.m.Load(all[i].key); ok {
					_ = v.(*entry).tenant.Close()
					c.m.Delete(all[i].key)
					c.log.Printf("tenant %s evicted (LRU pressure)", all[i].key)
					metrics.TenantEvictTotal.Inc()
					metrics.ActiveTenants.Dec()
				}
			}
		}
	}
}
