package discovery

import (
	"errors"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

// MultiServersDiscovery 不需要注册中心，服务列表由手工维护的服务发现的结构体
type MultiServersDiscovery struct {
	r       *rand.Rand // 随机数生成
	mu      sync.RWMutex
	servers []string
	index   int // 记录 Round Robin 算法已经轮询到的位置
}

var _ Discovery = (*MultiServersDiscovery)(nil)

func NewMultiServerDiscovery(servers []string) *MultiServersDiscovery {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, rpcAddr := range servers {
		log.Println("[discovery] add server:" + rpcAddr)
	}
	return &MultiServersDiscovery{
		r:       r,
		index:   r.Intn(math.MaxInt32 - 1),
		servers: servers,
	}
}

func (m *MultiServersDiscovery) Refresh() error {
	log.Println("[refresh] 当前实现不支持动态刷新")
	return nil
}

func (m *MultiServersDiscovery) Update(servers []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.servers = servers
	return nil
}

func (m *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	n := len(m.servers)
	if n == 0 {
		return "", errors.New("[rpc discovery]: no arvailable servers")
	}

	switch mode {
	case RandomMode:
		return m.servers[m.r.Intn(n)], nil
	case RoundRobinMode:
		s := m.servers[m.index%n]
		m.index = (m.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: not supported select mode")
	}
}

func (m *MultiServersDiscovery) GetAll() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	servers := make([]string, len(m.servers), len(m.servers))
	copy(servers, m.servers)

	return servers, nil
}
