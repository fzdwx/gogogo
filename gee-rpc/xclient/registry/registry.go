package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	DefaultRegisterPath  = "/rpc_/registry"
	DefaultHeadServerKey = "X-Rpc-Servers"
	defaultTimeout       = time.Minute * 5
)

type (
	// Register 是一个简单的注册中心，提供以下功能。
	// 添加服务器并接收心跳以使其保持活动状态。
	// 返回所有活动服务器并同时删除死服务器同步。
	Register struct {
		timeout time.Duration // timeout 服务器心跳超时时间
		servers map[string]*ServerItem
		mu      sync.Mutex
	}

	// ServerItem 代表每一个注册到注册中心的服务
	ServerItem struct {
		Addr  string
		start time.Time
	}
)

// addServer 添加服务实例，如果服务已经存在，则更新 start。
func (r *Register) addServer(addr string) {
	log.Printf("[ registry ] add : %s\n", addr)

	r.mu.Lock()
	defer r.mu.Unlock()

	server, ok := r.servers[addr]
	if ok {
		server.start = time.Now() // if exists, update start time to keep alive
	} else {
		r.servers[addr] = &ServerItem{
			Addr:  addr,
			start: time.Now(),
		}
	}
}

// aliveServers 返回可用的服务列表，如果存在超时的服务，则删除。
func (r *Register) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

func (r *Register) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		// keep it simple, server is in req.Header
		w.Header().Set(DefaultHeadServerKey, strings.Join(r.aliveServers(), ","))
	case http.MethodPost:
		// keep it simple, server is in req.Header
		addr := req.Header.Get(DefaultHeadServerKey)
		if addr == "" { // 服务端没有发送自己的地址
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.addServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// HandleHTTP 开始提供http服务，默认 /rpc_/registry
func HandleHTTP() {
	DefaultGeeRegister.HandleHTTP(DefaultRegisterPath)
}

// HandleHTTP registers an HTTP handler for registry messages on registryPath
//
// eg:	DefaultRegisterPath    = "/rpc_/registry"
func (r *Register) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("[ rpc registry ] start provide services:", registryPath)
}

var DefaultGeeRegister = NewRegister(defaultTimeout)

func NewRegister(timeout time.Duration) *Register {
	return &Register{
		timeout: timeout,
		servers: make(map[string]*ServerItem),
	}
}

// Heartbeat 向注册中心发送心跳
// cycleTime: 周期
func Heartbeat(registry, addr string, cycleTime time.Duration) {
	if cycleTime == 0 {
		cycleTime = defaultTimeout - time.Minute
	}
	err := sendHeartbeat(registry, addr)

	go func() {
		t := time.NewTicker(cycleTime)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registry, addr)
		}
	}()
}

func sendHeartbeat(registry string, addr string) error {
	log.Println("[ heartbeat ]", addr, "send to registry", registry)
	request, _ := http.NewRequest(http.MethodPost, registry, nil)
	request.Header.Set(DefaultHeadServerKey, addr)
	if _, err := http.DefaultClient.Do(request); err != nil {
		log.Println("[ heartbeat ] err:", err)
		return err
	}
	return nil
}
