package loadbalancer

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type IBackend interface {
	SetAlive(bool)
	IsAlive() bool
	GetURL() *url.URL
	GetActiveConnections() int
	Serve(http.ResponseWriter, *http.Request)
}

type Backend struct {
	url          *url.URL
	mux          *sync.RWMutex
	connections  int
	alive        bool
	reverseProxy *httputil.ReverseProxy
}

func (backend *Backend) SetAlive(isAlive bool) {
	backend.mux.Lock()
	backend.alive = isAlive
	backend.mux.Unlock()
}

func (backend *Backend) IsAlive() bool {
	backend.mux.RLock()
	defer backend.mux.RUnlock()
	return backend.alive
}

func (backend *Backend) GetURL() *url.URL {
	return backend.url
}

func (backend *Backend) GetActiveConnections() int {
	backend.mux.RLock()
	defer backend.mux.RUnlock()
	return backend.connections
}

func (backend *Backend) Serve(rw http.ResponseWriter, rq *http.Request) {
	defer func() {
		backend.mux.Lock()
		backend.connections--
		backend.mux.Unlock()
	}()

	backend.mux.Lock()
	backend.connections++
	backend.mux.Unlock()

	backend.reverseProxy.ServeHTTP(rw, rq)
}

func NewBackend(u *url.URL) *Backend {
	return &Backend{
		url:          u,
		alive:        true,
		mux:          &sync.RWMutex{},
		connections:  0,
		reverseProxy: httputil.NewSingleHostReverseProxy(u),
	}
}
