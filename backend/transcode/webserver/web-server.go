package webserver

import (
	"context"
	"net/http"
	"path/filepath"
	"sen1or/letslive/transcode/pkg/logger"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type WebServer struct {
	httpServer      *http.Server
	ListenPort      int
	AllowedSuffixes []string
	BaseDirectory   string
}

func NewWebServer(listenPort int, allowedSuffixes []string, baseDirectory string) *WebServer {
	return &WebServer{
		ListenPort:      listenPort,
		AllowedSuffixes: allowedSuffixes,
		BaseDirectory:   baseDirectory,
	}
}

func sanitizeRequestPath(requestPath string) string {
	trimmedPath := strings.TrimSpace(requestPath)
	cleanedRequestPath := filepath.Clean(trimmedPath)
	return cleanedRequestPath
}

func (ws *WebServer) ListenAndServe() {
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(ws.BaseDirectory))))
	router.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Use(corsMiddleware)

	ws.httpServer = &http.Server{
		Addr:         ":" + strconv.Itoa(ws.ListenPort),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := ws.httpServer.ListenAndServe(); err != nil {
		logger.Errorf("failed to start web server: %s", err.Error())
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Cache,Cache-Control,Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// shutdown gracefully shuts down the server without interrupting active connections.
func (ws *WebServer) Shutdown(ctx context.Context) error {
	if ws.httpServer == nil {
		logger.Warnf("web server instance not found, cannot shutdown.")
		return nil
	}

	logger.Infof("attempting graceful shutdown of web server...")
	err := ws.httpServer.Shutdown(ctx)
	if err != nil {
		logger.Errorf("web server shutdown failed: %v", err)
		return err
	}

	logger.Infof("web server shutdown completed.")
	return nil
}
