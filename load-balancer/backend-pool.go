package loadbalancer

import "errors"

type IBackendPool interface {
	GetBackends() []Backend
	GetNextBackend() (*Backend, error)
	AddBackend(Backend)
	GetPoolSize() int
}

type BackendPool struct {
	backends []Backend
}

func (lb *BackendPool) GetBackends() []Backend {
	return lb.backends
}

// TODO: check for eternal loops, set timeout etc...
func (lb *BackendPool) GetNextBackend() (*Backend, error) {
	var leastConnectedBackend *Backend = &lb.backends[0]

	for _, be := range lb.backends {
		if be.IsAlive() && be.GetActiveConnections() < leastConnectedBackend.GetActiveConnections() {
			leastConnectedBackend = &be
		}
	}

	if !leastConnectedBackend.IsAlive() {
		return nil, errors.New("no backend found")
	}

	return leastConnectedBackend, nil
}

func (lb *BackendPool) GetPoolSize() int {
	return len(lb.backends)
}

func (lb *BackendPool) AddBackend(be Backend) {
	lb.backends = append(lb.backends, be)
}

func NewBackendPool(backends []Backend) *BackendPool {
	return &BackendPool{
		backends: backends,
	}
}
