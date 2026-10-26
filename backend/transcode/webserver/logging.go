package webserver

import (
	"net/http"
	"strings"
	"time"

	"sen1or/letslive/shared/middlewares"
	"sen1or/letslive/shared/pkg/logger"
)

const (
	livePrefix = "/static/"
	vodPrefix  = "/static/vods/"
)

// isLivePath reports whether path is routed to the live LL-HLS handler.
func isLivePath(path string) bool {
	return strings.HasPrefix(path, livePrefix) && !strings.HasPrefix(path, vodPrefix)
}

// requestLogging logs requests through the shared LoggingMiddleware, except
// live LL-HLS requests: each viewer makes several per second (blocking
// playlist reloads and part fetches), so only their failures are logged.
func requestLogging(next http.Handler) http.Handler {
	shared := middlewares.LoggingMiddleware(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLivePath(r.URL.Path) {
			shared.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		if sw.status < http.StatusBadRequest {
			return
		}
		logger.Errorw(r.Context(), "failed api call",
			"requestId", r.Context().Value("requestId"),
			"duration", time.Since(start).Milliseconds(),
			"method", r.Method,
			"response#status", sw.status,
			"uri", r.RequestURI,
		)
	})
}

// statusWriter records the response status. The status stays 200 when the
// handler writes a body without calling WriteHeader.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status, w.wroteHeader = code, true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
