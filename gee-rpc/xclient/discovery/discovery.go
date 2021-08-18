package discovery

// SelectMode 负载均衡模式
type SelectMode int

// SelectMode 常量
const (
	RandomMode     = 1
	RoundRobinMode = 2
)

// Discovery 服务注册的接口定义
type Discovery interface {
	Refresh() error                      // Refresh 从注册中心更新服务列表
	Update(servers []string) error       // Update 手动更新服务列表
	Get(mode SelectMode) (string, error) // Get  根据负载均衡策略选择一个服务
	GetAll() ([]string, error)           // GetAll 返回所有服务信息
}
