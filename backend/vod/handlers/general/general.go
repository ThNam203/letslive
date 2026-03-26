package general

import (
	"encoding/json"
	"net/http"
	"sen1or/letslive/vod/handlers/basehandler"
	"sen1or/letslive/vod/response"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GeneralHandler struct {
	basehandler.BaseHandler
	DB *pgxpool.Pool
}

func NewGeneralHandler(db *pgxpool.Pool) *GeneralHandler {
	return &GeneralHandler{DB: db}
}

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

func (h *GeneralHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	h.WriteResponse(w, r.Context(), response.NewResponseFromTemplate[any](response.RES_ERR_ROUTE_NOT_FOUND, nil, nil, nil))
}
