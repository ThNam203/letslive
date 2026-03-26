package handlers

import (
	"encoding/json"
	"net/http"
	serviceresponse "sen1or/letslive/auth/response"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GeneralHandler struct {
	db *pgxpool.Pool
}

func NewGeneralHandler(db *pgxpool.Pool) *GeneralHandler {
	return &GeneralHandler{db: db}
}

func (h GeneralHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, r.Context(), serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_ROUTE_NOT_FOUND, nil, nil, nil))
}

func (h *GeneralHandler) RouteServiceHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := "ok"
	dbStatus := "ok"

	if err := h.db.Ping(r.Context()); err != nil {
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
