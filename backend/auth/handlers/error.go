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

type HTTPErrorResponse struct {
	StatusCode int    `json:"statusCode" example:"500"`
	Message    string `json:"message" example:"internal server error"`
}

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
