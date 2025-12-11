package utils

import (
	"context"
	"net/http"
	"sen1or/letslive/livestream/handlers/basehandler"
	"sen1or/letslive/livestream/response"
)

// WriteResponse is kept for backward compatibility but delegates to BaseHandler
func WriteResponse[T any](w http.ResponseWriter, ctx context.Context, res *response.Response[T]) {
	base := &basehandler.BaseHandler{}
	base.WriteResponse(w, ctx, res)
}
