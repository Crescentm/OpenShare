package cache

import (
	"sync"
	"time"
)

type TTL[K comparable, V any] struct {
	mu    sync.RWMutex
	ttl   time.Duration
	items map[K]ttlItem[V]
}

type ttlItem[V any] struct {
	value     V
	expiresAt time.Time
}

func NewTTL[K comparable, V any](ttl time.Duration) *TTL[K, V] {
	return &TTL[K, V]{
		ttl:   ttl,
		items: make(map[K]ttlItem[V]),
	}
}

func (c *TTL[K, V]) Get(key K) (V, bool) {
	var zero V
	if c == nil || c.ttl <= 0 {
		return zero, false
	}

	now := time.Now()
	c.mu.RLock()
	item, ok := c.items[key]
	if ok && now.Before(item.expiresAt) {
		c.mu.RUnlock()
		return item.value, true
	}
	c.mu.RUnlock()

	if ok {
		c.mu.Lock()
		if item, ok := c.items[key]; ok && !now.Before(item.expiresAt) {
			delete(c.items, key)
		}
		c.mu.Unlock()
	}

	return zero, false
}

func (c *TTL[K, V]) Set(key K, value V) {
	if c == nil || c.ttl <= 0 {
		return
	}

	c.mu.Lock()
	c.items[key] = ttlItem[V]{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

func (c *TTL[K, V]) Delete(key K) {
	if c == nil {
		return
	}

	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}
