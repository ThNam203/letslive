package middlewares

import (
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type LoggingMiddleware struct {
	logger *zap.Logger
}

func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

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

func (m *LoggingMiddleware) GetMiddleware(next http.Handler) http.Handler {
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

		fields := []zap.Field{
			zap.Int64("duration", duration),
			zap.String("method", r.Method),
			zap.String("remote#addr", remoteAddr),
			zap.Int("response#bytes", lrw.bytes),
			zap.Int("response#status", lrw.statusCode),
			zap.String("uri", r.RequestURI),
		}

		if lrw.statusCode/100 == 2 {
			m.logger.Info("success api call", fields...)
		} else {
			err := lrw.w.Header().Get("X-LetsLive-Error")
			if len(err) == 0 {
				m.logger.Info("failed api call", fields...)
			} else {
				m.logger.Error("failed api call: "+err, fields...)

			}
		}

		// TODO: prometheus
	})
}
