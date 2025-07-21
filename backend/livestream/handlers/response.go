package handlers

import (
	"encoding/json"
	"net/http"
	serviceresponse "sen1or/letslive/livestream/responses"
)

type ResponseHandler struct{}

func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

type serviceResponse struct {
	Message string `json:"message" example:"internal server error"`
	Data    any    `json:"data,omitempty"`
}

func (h *ResponseHandler) WriteErrorResponse(w http.ResponseWriter, err *serviceresponse.ServiceErrorResponse) {
	w.Header().Add("X-LetsLive-Error", err.Error())
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(serviceResponse{
		Message: err.Error(),
	})
}

func (h *ResponseHandler) WriteSuccessResponse(w http.ResponseWriter, success *serviceresponse.ServiceSuccessResponse, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(success.StatusCode)
	json.NewEncoder(w).Encode(serviceResponse{
		Message: success.Message,
		Data:    data,
	})
}

func (h *ResponseHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	h.WriteErrorResponse(w, serviceresponse.ErrRouteNotFound)
}
