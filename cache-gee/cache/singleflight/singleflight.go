package singleflight

import "sync"

// call 正在进行或者已经结束的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mtx sync.Mutex
	m   map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mtx.Lock()

	// init call map
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mtx.Unlock()
		c.wg.Wait()         // 如果请求正在进行中，则等待
		return c.val, c.err // 请求结束，返回结果
	}

	c := new(call)
	c.wg.Add(1)  // 发起请求前加锁
	g.m[key] = c // 添加到 g.m，表明 key 已经有对应的请求在处理
	g.mtx.Unlock()

	c.val, c.err = fn() // 调用 fn，发起请求
	c.wg.Done()         // 请求结束

	g.mtx.Lock()
	delete(g.m, key)
	g.mtx.Unlock()

	return c.val, c.err // 返回结果
}
