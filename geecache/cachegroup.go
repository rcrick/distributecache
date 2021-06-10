package geecache

import (
	"fmt"
	"log"
	"sync"
)

type CacheGroup struct {
	name   string
	cache  *cache
	getter Getter
	peers  PeerPicker
}

// for callbacl if not in cache, and load from local
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*CacheGroup)
)

func NewGroup(name string, maxBytes int64, getter Getter) *CacheGroup {
	if getter == nil {
		panic("getter is nil")
	}
	g := &CacheGroup{
		name:   name,
		cache:  &cache{maxBytes: maxBytes},
		getter: getter,
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *CacheGroup {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *CacheGroup) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is empty")
	}
	if v, ok := g.cache.get(key); ok {
		log.Println("[Cache] hit key: ", key)
		return v, nil
	}
	return g.load(key)
}
func (g *CacheGroup) load(key string) (value ByteView, err error) {
	fmt.Print("=======peers=", g.peers)
	if peer, ok := g.peers.PickPeer(key); ok {
		if value, err = g.getFromPeer(peer, key); err == nil {
			return value, nil
		}
		log.Println("[GeeCache] Failed to get from peer", err)
	}
	return g.getLocally(key)
}

func (g *CacheGroup) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	fmt.Print("========", peer)
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
func (g *CacheGroup) getLocally(key string) (ByteView, error) {
	v, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: v}, nil
}

func (g *CacheGroup) addToCache(key string, value ByteView) {
	g.cache.add(key, value)
}

func (g *CacheGroup) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("peer already exist")
	}
	g.peers = peers
}
