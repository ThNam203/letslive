package loadbalancer

type LBConfig struct {
	Name string `yaml:"name"`
	From string `yaml:"from"`
	To   []string
}

type TCPLoadBalancer struct {
	backendPool BackendPool
	config      LBConfig
}

type HTTPLoadBalancer struct {
	backendPool BackendPool
	config      LBConfig
}
