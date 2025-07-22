package middlewares

import (
	"context"
	"net/http"

	"github.com/gofrs/uuid/v5"
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

		ctx := context.WithValue(r.Context(), "requestId", requestId)
		r = r.WithContext(ctx)

		w.Header().Set("X-Request-ID", requestId)
		next.ServeHTTP(w, r)
	})
}
