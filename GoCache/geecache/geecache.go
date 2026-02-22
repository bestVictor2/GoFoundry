package geecache

import (
	"GoCache/geecache/singleflight"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name       string
	getter     Getter
	mainCache  cache
	peers      PeerPicker
	loader     *singleflight.Group
	defaultTTL time.Duration
	stats      groupStats
}

type groupStats struct {
	hits         uint64
	misses       uint64
	loads        uint64
	peerLoads    uint64
	peerFailures uint64
	localLoads   uint64
}

type Stats struct {
	Name         string `json:"name"`
	Hits         uint64 `json:"hits"`
	Misses       uint64 `json:"misses"`
	Loads        uint64 `json:"loads"`
	PeerLoads    uint64 `json:"peer_loads"`
	PeerFailures uint64 `json:"peer_failures"`
	LocalLoads   uint64 `json:"local_loads"`
	Entries      int    `json:"entries"`
	CacheBytes   int64  `json:"cache_bytes"`
	Evictions    uint64 `json:"evictions"`
}

type Option func(*Group)

func WithTTL(ttl time.Duration) Option {
	return func(g *Group) {
		g.defaultTTL = ttl
	}
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, getter Getter, cacheSize int64) *Group {
	return NewGroupWithOptions(name, getter, cacheSize)
}

func NewGroupWithOptions(name string, getter Getter, cacheSize int64, opts ...Option) *Group {
	if getter == nil {
		panic("geecache: getter is nil")
	}
	mu.Lock()
	defer mu.Unlock()
	group := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheSize},
		loader:    &singleflight.Group{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(group)
		}
	}
	groups[name] = group
	return group
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("geecache: key is empty")
	}
	if v, ok := g.mainCache.get(key); ok {
		atomic.AddUint64(&g.stats.hits, 1)
		return v, nil
	}
	atomic.AddUint64(&g.stats.misses, 1)
	return g.load(key)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("geecache: peer picker already registered")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	view, err := g.loader.Do(key, func() (interface{}, error) {
		atomic.AddUint64(&g.stats.loads, 1)
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if peerValue, peerErr := g.getFromPeer(peer, key); peerErr == nil {
					atomic.AddUint64(&g.stats.peerLoads, 1)
					g.populateCache(key, peerValue)
					return peerValue, nil
				}
				atomic.AddUint64(&g.stats.peerFailures, 1)
				log.Printf("geecache: failed to get from peer for key=%s", key)
			}
		}
		return g.getLocally(key)
	})
	if err != nil {
		return ByteView{}, err
	}
	return view.(ByteView), nil
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	atomic.AddUint64(&g.stats.localLoads, 1)
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, errors.New("geecache: key not found")
	}
	value := ByteView{b: cloneByte(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value, g.defaultTTL)
}

func (g *Group) Remove(key string) {
	g.mainCache.remove(key)
}

func (g *Group) RemoveMany(keys ...string) {
	for _, key := range keys {
		g.mainCache.remove(key)
	}
}

func (g *Group) GetMany(keys ...string) (map[string]ByteView, map[string]error) {
	values := make(map[string]ByteView, len(keys))
	errs := make(map[string]error)
	for _, key := range keys {
		value, err := g.Get(key)
		if err != nil {
			errs[key] = err
			continue
		}
		values[key] = value
	}
	if len(errs) == 0 {
		errs = nil
	}
	return values, errs
}

func (g *Group) Stats() Stats {
	entries, cacheBytes, evictions := g.mainCache.stats()
	return Stats{
		Name:         g.name,
		Hits:         atomic.LoadUint64(&g.stats.hits),
		Misses:       atomic.LoadUint64(&g.stats.misses),
		Loads:        atomic.LoadUint64(&g.stats.loads),
		PeerLoads:    atomic.LoadUint64(&g.stats.peerLoads),
		PeerFailures: atomic.LoadUint64(&g.stats.peerFailures),
		LocalLoads:   atomic.LoadUint64(&g.stats.localLoads),
		Entries:      entries,
		CacheBytes:   cacheBytes,
		Evictions:    evictions,
	}
}
