package gee

import (
	"net/http"
)

// HandlerFunc 请求处理器  给用户去实现
type HandlerFunc func(ctx *Context)

// Engine 实现 ServerHttp
type Engine struct {
	router *router
}

// New 构造函数 初始化router
func New() *Engine {
	return &Engine{router: newRouter()}
}

// GET 添加一个GET方式的请求处理器
func (e Engine) GET(path string, handler HandlerFunc) *Engine {
	e.router.addRouter("GET", path, handler)
	return &e
}

// POST 添加一个POST方式的请求处理器
func (e Engine) POST(path string, handler HandlerFunc) *Engine {
	e.router.addRouter("POST", path, handler)
	return &e
}

// Run 启动http服务器
func (e Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

// ServeHTTP 每次请求进来都调用这个方法进行处理
func (e Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 每次请求都new一个上下文
	c := newContext(w, req)
	// 然后进行处理
	e.router.handle(c)
}

// 工具方法
func makeRouteKey(method string, path string) string {
	return method + "-" + path
}
