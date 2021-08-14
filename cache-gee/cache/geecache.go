package cache

import (
	"fmt"
	"log"
	"sync"

	pb "cache/cachepb"
	"cache/singleflight"
)

// Getter 根据key加载缓存数据
type Getter interface {
	// Get callback
	Get(key string) ([]byte, error)
}

// GetterFunc impl Getter
type GetterFunc func(key string) ([]byte, error)

// Get GetterFunc impl Get func
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 可以看成是一个命名空间
// eg:和用户相关的就保存在name=user的cache中，和密码有关的就保存在name=password的cache中
type Group struct {
	name      string // namespace
	getter    Getter // 缓存未命中时的操作
	mainCache cache  // 带锁的缓存
	peers     PeerPicker
	// use singleflight.Group to make sure that
	// each key is only fetched once
	loader *singleflight.Group
}

var (
	mtx    sync.RWMutex              // 读写锁
	groups = make(map[string]*Group) // key:groupName,value:Group
)

// NewGroup 返回一个新的Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mtx.Lock()
	defer mtx.Unlock()

	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
		},
		loader: &singleflight.Group{},
	}

	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mtx.RLock()
	group := groups[name]
	mtx.RUnlock()
	return group
}

// Get 根据key返回存储的数据
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// 存在cache
	if v, ok := g.mainCache.get(key); ok {
		log.Printf("[Cache Hit]  key : %v\n", key)
		return v, nil
	}

	// miss cache
	return g.load(key)
}

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}

	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// 调用传入的miss cache callback
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	v := ByteView{b: cloneBytes(bytes)}

	// save cache to group
	g.populateCache(key, v)

	return v, nil
}

func (g *Group) populateCache(key string, v ByteView) {
	g.mainCache.add(key, v)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
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
