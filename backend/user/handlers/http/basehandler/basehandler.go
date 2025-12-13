package basehandler

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
)

type BaseHandler struct{}

// WriteResponse writes a JSON response to the http.ResponseWriter
// accepts any response type and handles requestId and statusCode
func (b *BaseHandler) WriteResponse(w http.ResponseWriter, ctx context.Context, res any) {
	// use reflection to set requestId and get statusCode
	resValue := reflect.ValueOf(res).Elem()

	if ctxRequestId, ok := ctx.Value("requestId").(string); ok && len(ctxRequestId) > 0 {
		requestIdField := resValue.FieldByName("RequestId")
		if requestIdField.IsValid() && requestIdField.CanSet() {
			requestIdField.SetString(ctxRequestId)
		}
	}

	var statusCode int
	statusCodeField := resValue.FieldByName("StatusCode")
	if statusCodeField.IsValid() && statusCodeField.CanInt() {
		statusCode = int(statusCodeField.Int())
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if statusCode > 0 {
		w.WriteHeader(statusCode)
	}
	json.NewEncoder(w).Encode(res)
}
