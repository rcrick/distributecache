package geecache

import (
	"fmt"
	"geecache/singleflight"
	"log"
	"sync"

	pb "geecache/geecachepb"
)

type CacheGroup struct {
	name   string
	cache  *cache
	getter Getter
	peers  PeerPicker
	loader *singleflight.Group
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
		loader: &singleflight.Group{},
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
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return view.(ByteView), nil
	}
	return
}

func (g *CacheGroup) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	 err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
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
