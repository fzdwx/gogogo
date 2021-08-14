package cache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

/*
  http request pattern:
  /bathPath/{groupName}/{key}
*/

const defaultBasePath = "/distributed_cache/"
const pattern = "/bathPath/{groupName}/{key}"

type HttpPool struct {
	// eg: http://localhost:9999/distributed_cache/
	self     string // 记录自己的地址 ip:port or website url
	basePath string // 作为通信地址的开头，用于节点的访问
}

func NewHttpPool(self string) *HttpPool {
	return &HttpPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log info with server name
func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP impl http url handler
// only handler starWith HttpPool.basePath
func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method
	if !strings.HasPrefix(path, p.basePath) {
		panic("HttpPool serving unexpected path: " + path)
	}
	p.Log("%s %s", method, path)

	// GET /bathPath/{groupName}/{key}
	if strings.HasPrefix(path, p.basePath) && method == http.MethodGet {
		p.handlerGet(path, w, r)
	}
}

func (p *HttpPool) handlerGet(path string, w http.ResponseWriter, r *http.Request) {
	// see pattern
	parts := strings.SplitN(path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}
