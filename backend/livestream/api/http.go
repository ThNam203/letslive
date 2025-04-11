package api

import (
	"context"
	"fmt"
	"net/http"
	"sen1or/letslive/livestream/config"
	"sen1or/letslive/livestream/handlers"
	"sen1or/letslive/livestream/middlewares"
	"sen1or/letslive/livestream/pkg/logger"

	"time"

	"go.uber.org/zap"
)

type APIServer struct {
	httpServer *http.Server
	logger     *zap.SugaredLogger
	config     *config.Config

	errorHandler      *handlers.ErrorHandler
	healthHandler     *handlers.HealthHandler
	livestreamHandler *handlers.LivestreamHandler
}

func NewAPIServer(livestreamHandler *handlers.LivestreamHandler, cfg *config.Config) *APIServer {
	return &APIServer{
		logger: logger.Logger,
		config: cfg,

		errorHandler:      handlers.NewErrorHandler(),
		healthHandler:     handlers.NewHeathHandler(),
		livestreamHandler: livestreamHandler,
	}
}

func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()
	//TODO: change to query livestreams
	sm.HandleFunc("GET /v1/livestreams", a.livestreamHandler.GetLivestreamsOfUserPublicHandler)
	sm.HandleFunc("GET /v1/livestreams/{livestreamId}", a.livestreamHandler.GetLivestreamByIdPublicHandler)
	sm.HandleFunc("GET /v1/livestreamings", a.livestreamHandler.GetLivestreamingsPublicHandler)
	sm.HandleFunc("GET /v1/popular-vods", a.livestreamHandler.GetPopularVODsPublicHandler)
	sm.HandleFunc("GET /v1/is-streaming", a.livestreamHandler.CheckIsUserLivestreamingPublicHandler)

	sm.HandleFunc("GET /v1/livestreams/author", a.livestreamHandler.GetAllLivestreamsOfAuthorPrivateHandler)
	sm.HandleFunc("PATCH /v1/livestreams/{livestreamId}", a.livestreamHandler.UpdateLivestreamPrivateHandler)
	sm.HandleFunc("DELETE /v1/livestreams/{livestreamId}", a.livestreamHandler.DeleteLivestreamPrivateHandler)

	sm.HandleFunc("PUT /v1/internal/livestreams/{livestreamId}", a.livestreamHandler.UpdateLivestreamInternalHandler)
	sm.HandleFunc("POST /v1/internal/livestreams", a.livestreamHandler.CreateLivestreamInternalHandler)

	sm.HandleFunc("GET /v1/health", a.healthHandler.GetHealthyStateHandler)

	sm.HandleFunc("GET /", a.errorHandler.RouteNotFoundHandler)

	finalHandler := middlewares.LoggingMiddleware(sm)

	return finalHandler
}

// ListenAndServe sets up and runs the HTTP server.
// it blocks until the server is shut down or an error occurs.
// it returns http.ErrServerClosed on graceful shutdown, otherwise the error.
func (a *APIServer) ListenAndServe(useTLS bool) error { // Changed signature to return error
	addr := fmt.Sprintf("%s:%d", a.config.Service.APIBindAddress, a.config.Service.APIPort)

	a.httpServer = &http.Server{
		Addr:         addr,
		Handler:      a.getHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// start the server (this will block)
	var err error
	if useTLS {
		err = fmt.Errorf("TLS not implemented")
	} else {
		err = a.httpServer.ListenAndServe()
	}

	// This line is reached when ListenAndServe returns.
	// It returns http.ErrServerClosed if Shutdown was called gracefully.
	// Otherwise, it returns the error that caused it to stop.
	if err != nil && err != http.ErrServerClosed {
		logger.Errorf("server listener error: %v", err)
		return err
	}

	// If err is nil or http.ErrServerClosed, it means server stopped cleanly or via Shutdown.
	return nil
}

// shutdown gracefully shuts down the server without interrupting active connections.
func (a *APIServer) Shutdown(ctx context.Context) error {
	if a.httpServer == nil {
		logger.Warnf("server instance not found, cannot shutdown.")
		return nil
	}

	logger.Infof("attempting graceful shutdown of server...")
	err := a.httpServer.Shutdown(ctx)
	if err != nil {
		logger.Errorf("server shutdown failed: %v", err)
		return err
	}

	logger.Infof("server shutdown completed.")
	return nil
}
