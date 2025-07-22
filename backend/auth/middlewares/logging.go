package middlewares

import (
	"net"
	"net/http"
	"sen1or/letslive/auth/pkg/logger"
	"time"
)

type loggingResponseWriter struct {
	w          http.ResponseWriter
	statusCode int
	bytes      int
}

func (lrw *loggingResponseWriter) Header() http.Header {
	return lrw.w.Header()
}

func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	wb, err := lrw.w.Write(data)
	lrw.bytes += wb
	return wb, err
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.w.WriteHeader(statusCode)
	lrw.statusCode = statusCode
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()
		lrw := &loggingResponseWriter{w: w}
		next.ServeHTTP(lrw, r)

		duration := time.Since(timeStart).Milliseconds()
		remoteAddr := r.Header.Get("X-Forwarded-For")
		if remoteAddr == "" {
			if ip, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
				remoteAddr = "unknown address"
			} else {
				remoteAddr = ip
			}
		}

		requestId := r.Context().Value("requestId")

		fields := []any{
			"requestId", requestId,
			"duration", duration,
			"method", r.Method,
			"remote#addr", remoteAddr,
			"response#bytes", lrw.bytes,
			"response#status", lrw.statusCode,
			"uri", r.RequestURI,
		}

		if lrw.statusCode/100 == 2 {
			if r.RequestURI != "/v1/health" {
				logger.Infow("success api call", fields...)
			}
		} else {
			err := lrw.w.Header().Get("X-LetsLive-Error")
			if len(err) == 0 {
				logger.Infow("failed api call", fields...)
			} else {
				logger.Errorw("failed api call: "+err, fields...)
			}
		}
	})
}
