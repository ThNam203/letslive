package handlers

import (
	"encoding/json"
	"net/http"
	serviceresponse "sen1or/letslive/auth/responses"
)

type ResponseHandler struct{}

func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

type serviceResponse struct {
	Message string `json:"message" example:"internal server error"`
	Data    any    `json:"data,omitempty"`
}

func (h ResponseHandler) WriteErrorResponse(w http.ResponseWriter, err *serviceresponse.ServiceErrorResponse) {
	w.Header().Add("X-LetsLive-Error", err.Message)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(serviceResponse{
		Message: err.Message,
	})
}

func (h ResponseHandler) WriteSuccessResponse(w http.ResponseWriter, successRes *serviceresponse.ServiceSuccessResponse, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(successRes.StatusCode)
	json.NewEncoder(w).Encode(serviceResponse{
		Message: successRes.Message,
		Data:    data,
	})
}

func (h ResponseHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	h.WriteErrorResponse(w, serviceresponse.ErrRouteNotFound)
}
