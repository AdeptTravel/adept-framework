// internal/cache/lru.go
//
// Tiny LRU cache used by the view engine to store parsed *template.Template
// sets.  No external deps; good for a few thousand entries.
package cache

import "container/list"

// LRU is a non‑generic least‑recently‑used cache.
// Keys must be comparable; values can be any.
type LRU struct {
	cap  int
	ll   *list.List
	dict map[any]*list.Element
}

type pair struct {
	key any
	val any
}

// New returns an LRU with the given capacity.  Panics on cap < 1.
func New(capacity int) *LRU {
	if capacity < 1 {
		panic("cache: capacity must be ≥1")
	}
	return &LRU{
		cap:  capacity,
		ll:   list.New(),
		dict: make(map[any]*list.Element, capacity),
	}
}

// Get retrieves a value or nil and marks it MRU.
func (c *LRU) Get(key any) (val any, ok bool) {
	if ele, hit := c.dict[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(pair).val, true
	}
	return nil, false
}

// Add inserts or updates a value.
func (c *LRU) Add(key, val any) {
	if ele, hit := c.dict[key]; hit {
		ele.Value = pair{key, val}
		c.ll.MoveToFront(ele)
		return
	}
	ele := c.ll.PushFront(pair{key, val})
	c.dict[key] = ele
	if c.ll.Len() > c.cap {
		last := c.ll.Back()
		c.ll.Remove(last)
		delete(c.dict, last.Value.(pair).key)
	}
}

// Len reports current size.
func (c *LRU) Len() int { return c.ll.Len() }
