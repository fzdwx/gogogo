package cache

import (
	"fmt"
	"log"
	"sync"
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

func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
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
