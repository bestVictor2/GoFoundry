package geecache

import (
	"GoCache/lru"
	"sync"
	"time"
)

type cache struct {
	mu         sync.RWMutex
	lru        *lru.Cache
	cacheBytes int64
	expireAt   map[string]time.Time
	evictions  uint64
}

func (c *cache) ensureLRU() {
	if c.lru != nil {
		return
	}
	c.expireAt = make(map[string]time.Time)
	c.lru = lru.New(c.cacheBytes, func(key string, _ lru.Value) {
		delete(c.expireAt, key)
		c.evictions++
	})
}

func (c *cache) add(k string, value ByteView, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ensureLRU()
	c.lru.Add(k, value)
	if ttl > 0 {
		c.expireAt[k] = time.Now().Add(ttl)
		return
	}
	delete(c.expireAt, k)
}
func (c *cache) get(k string) (v ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if exp, hasExp := c.expireAt[k]; hasExp && time.Now().After(exp) {
		c.lru.RemoveKey(k)
		delete(c.expireAt, k)
		return ByteView{}, false
	}
	if cached, hit := c.lru.Get(k); hit {
		return cached.(ByteView), true
	}
	return ByteView{}, false
}
func (c *cache) remove(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	c.lru.RemoveKey(k)
	delete(c.expireAt, k)
}

func (c *cache) stats() (entries int, bytes int64, evictions uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.lru == nil {
		return 0, 0, c.evictions
	}
	return c.lru.Len(), c.lru.Bytes(), c.evictions
}
