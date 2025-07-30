package middlewares

import (
	"context"
	"net/http"
	"sen1or/letslive/user/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := r.Header.Get("X-Request-ID")

		if requestId == "" {
			requestObj, err := uuid.NewGen().NewV4()
			if err != nil {
				requestId = uuid.Nil.String()
			} else {
				requestId = requestObj.String()
			}
		}

		// add the request id to the span
		if span := trace.SpanFromContext(r.Context()); span != nil {
			span.SetAttributes(attribute.String("http.request_id", requestId))
		} else {
			logger.Warnf("missing span from context")
		}

		ctx := context.WithValue(r.Context(), "requestId", requestId)
		r = r.WithContext(ctx)

		w.Header().Set("X-Request-ID", requestId)
		next.ServeHTTP(w, r)
	})
}
