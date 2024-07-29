package loadbalancer

import "net/http"

type LoadBalancer struct {
	backendPool BackendPool
}

func (lb *LoadBalancer) Serve(rw http.ResponseWriter, rq *http.Request) {
	var be, err = lb.backendPool.GetNextBackend();
	if err != nil {
		http.Error(rw, "Load balancer is not working properly!", http.StatusInternalServerError)
		return;
	}

	be.Serve(rw, rq)
}