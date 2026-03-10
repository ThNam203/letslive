package general

import (
	"net/http"
	"sen1or/letslive/vod/handlers/basehandler"
	"sen1or/letslive/vod/response"
)

type GeneralHandler struct {
	basehandler.BaseHandler
}

func NewGeneralHandler() *GeneralHandler {
	return &GeneralHandler{}
}

func (h *GeneralHandler) RouteServiceHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *GeneralHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	h.WriteResponse(w, r.Context(), response.NewResponseFromTemplate[any](response.RES_ERR_ROUTE_NOT_FOUND, nil, nil, nil))
}
