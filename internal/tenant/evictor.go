// evictor.go houses the eviction loop for Cache.  Every EvictInterval it
// scans the map and removes:
//
//   - tenants idle longer than idleTTL
//   - least-recently-used tenants when map size exceeds maxEntries
//
// Each eviction event is logged and updates Prometheus counters.
package tenant

import (
	"sort"
	"sync/atomic"
	"time"

	"github.com/AdeptTravel/adept-framework/internal/metrics"
)

func (c *Cache) evictLoop() {
	for range c.evictTicker.C {
		now := time.Now().UnixNano()
		var count int

		// ----------------------------------------------------------------
		// Idle eviction pass
		// ----------------------------------------------------------------
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

		// ----------------------------------------------------------------
		// LRU eviction pass
		// ----------------------------------------------------------------
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
