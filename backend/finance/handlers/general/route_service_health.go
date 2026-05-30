package general

import (
	"encoding/json"
	"net/http"
)

func (h *GeneralHandler) RouteServiceHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := "ok"
	dbStatus := "ok"

	if err := h.DB.Ping(r.Context()); err != nil {
		status = "degraded"
		dbStatus = "unavailable"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(map[string]any{
		"status": status,
		"checks": map[string]string{
			"database": dbStatus,
		},
	})
}
