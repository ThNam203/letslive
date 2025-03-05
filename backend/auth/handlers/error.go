package handlers

import (
	"encoding/json"
	"net/http"
	servererrors "sen1or/lets-live/auth/errors"
)

type ErrorHandler struct{}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// Set the error message to the custom "X-LetsLive-Error" header
// The function doesn't end the request, if so use WriteErrorResponse
func (h *ErrorHandler) SetError(w http.ResponseWriter, err error) {
	w.Header().Add("X-LetsLive-Error", err.Error())
}

type HTTPErrorResponse struct {
	StatusCode int    `json:"statusCode" example:"500"`
	Message    string `json:"message" example:"internal server error"`
}

// Set error to the custom header and write the error to the request
// After calling, the request will end and no other write should be done
func (h *ErrorHandler) WriteErrorResponse(w http.ResponseWriter, err *servererrors.ServerError) {
	w.Header().Add("X-LetsLive-Error", err.Message)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(HTTPErrorResponse{
		Message:    err.Message,
		StatusCode: err.StatusCode,
	})
}

func (h *ErrorHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	h.WriteErrorResponse(w, servererrors.ErrRouteNotFound)
}
