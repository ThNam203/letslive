package handlers

import (
	"encoding/json"
	"net/http"
	"sen1or/letslive/user/response"
)

func writeResponse(w http.ResponseWriter, res *response.Response[any]) {
	res.RequestId = w.Header().Get("requestId")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(res.StatusCode)
	json.NewEncoder(w).Encode(res)
}
