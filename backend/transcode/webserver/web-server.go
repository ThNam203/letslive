package webserver

import (
	"context"
	"net/http"
	"sen1or/letslive/shared/middlewares"
	"sen1or/letslive/shared/pkg/logger"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// writeTimeout is long enough for a blocking playlist request to be held
// until the next part is ready.
const writeTimeout = 30 * time.Second

type WebServer struct {
	httpServer    *http.Server
	ListenPort    int
	BaseDirectory string
	live          http.Handler
}

// NewWebServer serves VODs from baseDirectory under /static/vods/ and
// everything else under /static/ through live, with the /static prefix
// removed.
func NewWebServer(listenPort int, baseDirectory string, live http.Handler) *WebServer {
	return &WebServer{
		ListenPort:    listenPort,
		BaseDirectory: baseDirectory,
		live:          live,
	}
}

func (ws *WebServer) routes() http.Handler {
	router := mux.NewRouter()
	router.PathPrefix(vodPrefix).Handler(http.StripPrefix(livePrefix, http.FileServer(http.Dir(ws.BaseDirectory))))
	router.PathPrefix(livePrefix).Handler(http.StripPrefix("/static", ws.live))
	router.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsMiddleware := middlewares.NewCORSMiddleware()
	router.Use(corsMiddleware.GetMiddleware)
	router.Use(middlewares.RequestIDMiddleware)
	router.Use(requestLogging)
	return router
}

func (ws *WebServer) ListenAndServe() {
	ws.httpServer = &http.Server{
		Addr:         ":" + strconv.Itoa(ws.ListenPort),
		Handler:      ws.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: writeTimeout,
	}

	if err := ws.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Errorf(context.TODO(), "failed to start web server: %s", err.Error())
	}
}

// shutdown gracefully shuts down the server without interrupting active connections.
func (ws *WebServer) Shutdown(ctx context.Context) error {
	if ws.httpServer == nil {
		logger.Warnf(ctx, "web server instance not found, cannot shutdown.")
		return nil
	}

	logger.Infof(ctx, "attempting graceful shutdown of web server...")
	err := ws.httpServer.Shutdown(ctx)
	if err != nil {
		logger.Errorf(ctx, "web server shutdown failed: %v", err)
		return err
	}

	logger.Infof(ctx, "web server shutdown completed.")
	return nil
}
