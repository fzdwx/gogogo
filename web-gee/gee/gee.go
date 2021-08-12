package gee

import (
	"fmt"
	"net/http"
)

// HandlerFunc 请求处理器  给用户去实现
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// Engine 实现 ServerHttp
type Engine struct {
	// 路由表 eg:GET-hello:handler
	routers map[string]HandlerFunc
}

// New 构造函数 初始化router
func New() *Engine {
	return &Engine{routers: make(map[string]HandlerFunc)}
}

// addRouter 添加路由
func (e Engine) addRouter(method, path string, handler HandlerFunc) *Engine {
	key := makeRouteKey(method, path)
	e.routers[key] = handler
	return &e
}

// GET 添加一个GET方式的请求处理器
func (e Engine) GET(path string, handler HandlerFunc) *Engine {
	e.addRouter("GET", path, handler)
	return &e
}

// POST 添加一个POST方式的请求处理器
func (e Engine) POST(path string, handler HandlerFunc) *Engine {
	e.addRouter("POST", path, handler)
	return &e
}

// Run 启动http服务器
func (e Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

func (e Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := makeRouteKey(req.Method, req.URL.Path)
	if handler, ok := e.routers[key]; ok {
		fmt.Println(ok)
		handler(w, req)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}

// 工具方法
func makeRouteKey(method string, path string) string {
	return method + "-" + path
}
