package gee

import (
	"net/http"
	"strings"
)

type router struct {
	// root节点，分为GET,POST...
	roots map[string]*node
	// 路由表
	// eg:GET-hello:handlerFunc
	// eg:POST-/person/:name/:age:handlerFunc
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// addRouter 添加路由
func (r *router) addRouter(method, path string, handler HandlerFunc) *router {
	parts := parsePath(path)
	key := makeRouteKey(method, path)

	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(path, parts, 0)
	r.handlers[key] = handler
	return r
}

func (r *router) getRouter(method string, path string) (*node, map[string]string) {
	searchPaths := parsePath(path)
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchPaths, 0)
	if n != nil {
		parts := parsePath(n.pattern)
		for i, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchPaths[i]
			}
			if path[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchPaths[i:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

// handle 处理当前请求，遍历路由表，找到对应的handlerFunc进行处理
func (r router) handle(c *Context) {

	n, params := r.getRouter(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := makeRouteKey(c.Method, n.pattern)
		// 路由匹配成功，调用用户实现的handlerFunc
		r.handlers[key](c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}

// 工具方法
func makeRouteKey(method string, path string) string {
	return method + "-" + path
}

func parsePath(pattern string) (parts []string) {
	vs := strings.Split(pattern, "/")
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}
