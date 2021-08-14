package cache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"cache/consistenthash"
)

/*
  http request pattern:
  /bathPath/{groupName}/{key}
*/

const (
	defaultBasePath = "/distributed_cache/"
	defaultReplicas = 50
	pattern         = "/bathPath/{groupName}/{key}"
)

type HttpPool struct {
	// eg: http://localhost:9999/distributed_cache/
	self        string // 记录自己的地址 ip:port or website url
	basePath    string // 作为通信地址的开头，用于节点的访问
	mtx         sync.Mutex
	peers       *consistenthash.Map    // 一致性哈希算法的 Map，用来根据具体的 key 选择节点。
	httpGetters map[string]*httpGetter // key: http://localhost:9999
}

func NewHttpPool(self string) *HttpPool {
	return &HttpPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (h *HttpPool) Set(peers ...string) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	h.peers = consistenthash.New(defaultReplicas, nil)
	h.peers.Add(peers...)
	h.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		h.httpGetters[peer] = &httpGetter{baseURL: peer + h.basePath}
	}
}

// PickPeer picks a peer according to key
func (h *HttpPool) PickPeer(key string) (peer PeerGetter, ok bool) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	if peer := h.peers.Get(key); peer != "" && peer != h.self {
		h.Log("Pick peer %s", peer)
		return h.httpGetters[peer], true
	}
	return nil, false
}

// Log info with server name
func (h *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", h.self, fmt.Sprintf(format, v...))
}

// ServeHTTP impl http url handler
// only handler starWith HttpPool.basePath
func (h *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method
	if !strings.HasPrefix(path, h.basePath) {
		panic("HttpPool serving unexpected path: " + path)
	}
	h.Log("%s %s", method, path)

	// GET /bathPath/{groupName}/{key}
	if strings.HasPrefix(path, h.basePath) && method == http.MethodGet {
		h.handlerGet(path, w, r)
	}
}

func (h *HttpPool) handlerGet(path string, w http.ResponseWriter, r *http.Request) {
	// see pattern
	parts := strings.SplitN(path[len(h.basePath):], "/", 2)
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

type httpGetter struct {
	baseURL string
}

// Get  用于从对应 group 查找缓存值。
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", resp.Status)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)
