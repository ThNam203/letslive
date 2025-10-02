package handlers

import (
	"net/http"
	"sen1or/letslive/user/response"
)

type GeneralHandler struct{}

func NewGeneralHandler() *GeneralHandler {
	return &GeneralHandler{}
}

func (h GeneralHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, r.Context(), response.NewResponseFromTemplate[any](response.RES_ERR_ROUTE_NOT_FOUND, nil, nil, nil))
}

func (h GeneralHandler) RouteServiceHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
