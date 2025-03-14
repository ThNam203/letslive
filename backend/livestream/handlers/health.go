package handlers

import "net/http"

type HealthHandler struct{}

func NewHeathHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) GetHealthyStateHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
