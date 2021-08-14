package cache

import (
	"sync"

	"cache/lru"
)

// cache 添加并发的锁控制，在Cache的继承上包装一层
type cache struct {
	mtx        sync.Mutex // 锁
	lru        *lru.Cache // lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mtx.Lock()
	defer c.mtx.Unlock() // 方法结束的时候解锁
	initLur(c)
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock() // 方法结束的时候解锁

	if c.lru == nil {
		return
	}

	// doGet
	v, ok := c.lru.Get(key)
	if ok {
		return v.(ByteView), ok
	}

	return
}

func initLur(c *cache) {
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
}
