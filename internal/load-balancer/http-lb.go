package loadbalancer

import (
	"log"
	"net/http"
	"net/url"

	config "sen1or/lets-live/internal/config"
)

type HTTPLoadBalancer struct {
	backendPool BackendPool
	config      config.LBSetting
}

func NewHTTPLoadBalancer(config config.LBSetting) *HTTPLoadBalancer {
	backends := make([]Backend, 0)

	for _, address := range config.To {
		url, err := url.Parse(address)
		if err != nil {
			log.Printf("backend address '%s' failed to parse", address)
			continue
		}
		be := NewBackend(url)
		backends = append(backends, *be)
	}

	if len(backends) == 0 {
		log.Panic("no backend found")
	}

	return &HTTPLoadBalancer{
		backendPool: *NewBackendPool(backends),
		config:      config,
	}
}

func (lb *HTTPLoadBalancer) ListenAndServe() {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", lb.serve)
	http.ListenAndServe(lb.config.From, serveMux)
}

func (lb *HTTPLoadBalancer) serve(rw http.ResponseWriter, rq *http.Request) {
	var be, err = lb.backendPool.GetNextBackend()

	if err != nil {
		http.Error(rw, "load balancer is not working properly!", http.StatusInternalServerError)
		return
	}

	be.Serve(rw, rq)
}
