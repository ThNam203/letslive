package middlewares

import "net/http"

// MaxBodySizeMiddleware limits the size of incoming request bodies to prevent
// resource exhaustion from arbitrarily large payloads. Individual handlers
// that need larger limits (e.g. file uploads) can override this by calling
// http.MaxBytesReader with their own limit.
func MaxBodySizeMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
