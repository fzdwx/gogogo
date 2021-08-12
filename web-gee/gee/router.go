package gee

import "fmt"

type router struct {
	// 路由表 eg:GET-hello:handler
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{handlers: make(map[string]HandlerFunc)}
}

// addRouter 添加路由
func (r router) addRouter(method, path string, handler HandlerFunc) *router {
	key := makeRouteKey(method, path)
	r.handlers[key] = handler
	return &r
}

// handle 处理当前请求，遍历路由表，找到对应的handlerFunc进行处理
func (r router) handle(c *Context) {
	key := makeRouteKey(c.Method, c.Path)
	if handler, ok := r.handlers[key]; ok {
		// 路由匹配成功，调用用户实现的handlerFunc
		handler(c)
	} else {
		fmt.Fprintf(c.Writer, "404 NOT FOUND: %s\n", c.Path)
	}
}
