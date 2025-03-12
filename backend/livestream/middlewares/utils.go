package middlewares

import (
	"encoding/json"
	"net/http"
	servererrors "sen1or/lets-live/livestream/errors"
)

func writeErrorResponse(w http.ResponseWriter, err *servererrors.ServerError) {
	type HTTPErrorResponse struct {
		StatusCode int    `json:"statusCode" example:"500"`
		Message    string `json:"message" example:"internal server error"`
	}

	w.Header().Add("X-LetsLive-Error", err.Error())
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(HTTPErrorResponse{
		Message:    err.Error(),
		StatusCode: err.StatusCode,
	})
}
