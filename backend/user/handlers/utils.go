package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/user/response"
)

func writeResponse(w http.ResponseWriter, ctx context.Context, res *response.Response[any]) {
	requestId, ok := ctx.Value("requestId").(string)
	if ok && len(requestId) > 0 {
		res.RequestId = requestId
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(res.StatusCode)
	json.NewEncoder(w).Encode(res)
}
