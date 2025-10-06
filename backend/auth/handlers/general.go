package handlers

import (
	"net/http"
	serviceresponse "sen1or/letslive/auth/response"
)

type GeneralHandler struct{}

func NewGeneralHandler() *GeneralHandler {
	return &GeneralHandler{}
}

func (h GeneralHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, r.Context(), serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_ROUTE_NOT_FOUND, nil, nil, nil))
}

func (h *GeneralHandler) RouteServiceHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
