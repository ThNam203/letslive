package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorHandler struct{}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// Set the error message to the custom "X-LetsLive-Error" header
// The function doesn't end the request, if so call errorResponse
func (h *ErrorHandler) SetError(w http.ResponseWriter, err error) {
	w.Header().Add("X-LetsLive-Error", err.Error())
}

type HTTPErrorResponse struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"internal server error"`
}

// Set error to the custom header and write the error to the request
// After calling, the request will end and no other write should be done
func (h *ErrorHandler) WriteErrorResponse(w http.ResponseWriter, status int, err error) {
	w.Header().Add("X-LetsLive-Error", err.Error())
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(HTTPErrorResponse{
		Message: err.Error(),
		Code:    status,
	})
}

func (h *ErrorHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	h.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("route not found"))
}
