package tenant

import (
	"sort"
	"sync/atomic"
	"time"

	"github.com/AdeptTravel/adept-framework/internal/metrics"
)

// evictLoop runs in a goroutine.  Every EvictInterval it removes idle
// tenants and trims the cache to MaxEntries via a simple LRU pass.
func (c *Cache) evictLoop() {
	for range c.evictTicker.C {
		now := time.Now().UnixNano()
		var count int

		// First pass: idle eviction.
		c.m.Range(func(key, value any) bool {
			count++
			ent := value.(*entry)
			idle := time.Duration(now-atomic.LoadInt64(&ent.lastSeen)) * time.Nanosecond
			if idle > c.idleTTL {
				_ = ent.tenant.Close()
				c.m.Delete(key)
				metrics.TenantEvictTotal.Inc()
				metrics.ActiveTenants.Dec()
			}
			return true
		})

		// Second pass: LRU eviction if over capacity.
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
					metrics.TenantEvictTotal.Inc()
					metrics.ActiveTenants.Dec()
				}
			}
		}
	}
}
