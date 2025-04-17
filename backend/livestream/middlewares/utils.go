package middlewares

import (
	"encoding/json"
	"net/http"
	serviceresponse "sen1or/letslive/livestream/responses"
)

func writeErrorResponse(w http.ResponseWriter, err *serviceresponse.ServiceErrorResponse) {
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
