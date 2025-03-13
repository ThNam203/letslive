package handlers

import (
	"encoding/json"
	"net/http"
	servererrors "sen1or/letslive/user/errors"
)

type ErrorHandler struct{}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

type HTTPErrorResponse struct {
	StatusCode int    `json:"statusCode" example:"500"`
	Message    string `json:"message" example:"internal server error"`
}

// Set error to the custom header and write the error to the request
// After calling, the request will end and no other write should be done
func (h *ErrorHandler) WriteErrorResponse(w http.ResponseWriter, err *servererrors.ServerError) {
	w.Header().Add("X-LetsLive-Error", err.Error())
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(HTTPErrorResponse{
		Message:    err.Error(),
		StatusCode: err.StatusCode,
	})
}

func (h *ErrorHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	h.WriteErrorResponse(w, servererrors.ErrRouteNotFound)
}
