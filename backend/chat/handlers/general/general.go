package general

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"sen1or/letslive/chat/handlers/basehandler"
	"sen1or/letslive/chat/response"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const healthPingTimeout = 2 * time.Second

type GeneralHandler struct {
	basehandler.BaseHandler
	mongoClient     *mongo.Client
	isNATSConnected func() bool
}

func NewGeneralHandler(mongoClient *mongo.Client, isNATSConnected func() bool) *GeneralHandler {
	return &GeneralHandler{mongoClient: mongoClient, isNATSConnected: isNATSConnected}
}

func (h *GeneralHandler) RouteServiceHealth(w http.ResponseWriter, r *http.Request) {
	status, code := "ok", http.StatusOK
	checks := map[string]string{"database": "ok", "nats": "ok"}

	ctx, cancel := context.WithTimeout(r.Context(), healthPingTimeout)
	defer cancel()
	if err := h.mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		checks["database"] = "unavailable"
		status, code = "degraded", http.StatusServiceUnavailable
	}
	if !h.isNATSConnected() {
		checks["nats"] = "unavailable"
		status, code = "degraded", http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": status, "checks": checks})
}

func (h *GeneralHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	h.WriteResponse(w, r.Context(), response.NewResponseFromTemplate[any](response.RES_ERR_ROUTE_NOT_FOUND, nil, nil, nil))
}
