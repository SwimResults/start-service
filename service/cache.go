package service

import (
	"sync"
	"time"
)

// ttlCache caches values per key for a given time. On a failed reload it keeps serving the stale value.
type ttlCache[K comparable, V any] struct {
	mu      sync.Mutex
	entries map[K]ttlEntry[V]
	now     func() time.Time
}

type ttlEntry[V any] struct {
	value    V
	loadedAt time.Time
}

func newTTLCache[K comparable, V any]() *ttlCache[K, V] {
	return &ttlCache[K, V]{entries: map[K]ttlEntry[V]{}, now: time.Now}
}

// get returns the cached value if younger than ttl, else calls load. Concurrent misses may load twice;
// that's fine for the short TTLs used here.
func (c *ttlCache[K, V]) get(key K, ttl time.Duration, load func() (V, error)) (V, error) {
	c.mu.Lock()
	e, ok := c.entries[key]
	now := c.now()
	c.mu.Unlock()
	if ok && now.Sub(e.loadedAt) < ttl {
		return e.value, nil
	}

	v, err := load()
	if err != nil {
		if ok {
			return e.value, nil
		}
		return v, err
	}
	c.mu.Lock()
	c.entries[key] = ttlEntry[V]{value: v, loadedAt: now}
	c.mu.Unlock()
	return v, nil
}

func (c *ttlCache[K, V]) delete(key K) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}
