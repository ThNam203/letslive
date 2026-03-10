package api

import (
	"context"
	"fmt"
	"net/http"
	"sen1or/letslive/livestream/config"
	"sen1or/letslive/livestream/handlers/general"
	"sen1or/letslive/livestream/handlers/livestream"
	"sen1or/letslive/livestream/middlewares"
	"sen1or/letslive/livestream/pkg/logger"

	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type APIServer struct {
	httpServer *http.Server
	logger     *zap.SugaredLogger
	config     *config.Config

	generalHandler    *general.GeneralHandler
	livestreamHandler *livestream.LivestreamHandler
}

func NewAPIServer(livestreamHandler *livestream.LivestreamHandler, cfg *config.Config) *APIServer {
	return &APIServer{
		logger: logger.Logger,
		config: cfg,

		generalHandler:    general.NewGeneralHandler(),
		livestreamHandler: livestreamHandler,
	}
}

func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()

	wrap := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		sm.Handle(pattern, http.HandlerFunc(handlerFunc))
	}

	wrap("GET /v1/popular-livestreams", a.livestreamHandler.GetRecommendedLivestreamsPublicHandler)
	wrap("GET /v1/livestreams", a.livestreamHandler.GetLivestreamOfUserPublicHandler)

	wrap("POST /v1/internal/livestreams/{livestreamId}/end", a.livestreamHandler.EndLivestreamAndCreateVODInternalHandler)
	wrap("POST /v1/internal/livestreams", a.livestreamHandler.CreateLivestreamInternalHandler)

	wrap("GET /v1/health", a.generalHandler.RouteServiceHealth)

	wrap("GET /", a.generalHandler.RouteNotFoundHandler)

	// TODO: remove filter
	finalHandler := otelhttp.NewHandler(sm, "/", otelhttp.WithFilter(func(r *http.Request) bool {
		return r.URL.Path != "/v1/health" // exclude this path from tracing
	}))
	finalHandler = middlewares.LoggingMiddleware(finalHandler)
	finalHandler = middlewares.RequestIDMiddleware(finalHandler)

	return finalHandler
}

// ListenAndServe sets up and runs the HTTP server.
// it blocks until the server is shut down or an error occurs.
// it returns http.ErrServerClosed on graceful shutdown, otherwise the error.
func (a *APIServer) ListenAndServe(ctx context.Context, useTLS bool) error { // Changed signature to return error
	addr := fmt.Sprintf("%s:%d", a.config.Service.APIBindAddress, a.config.Service.APIPort)

	a.httpServer = &http.Server{
		Addr:         addr,
		Handler:      a.getHandler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
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
		logger.Errorf(ctx, "server listener error: %v", err)
		return err
	}

	// If err is nil or http.ErrServerClosed, it means server stopped cleanly or via Shutdown.
	return nil
}

// shutdown gracefully shuts down the server without interrupting active connections.
func (a *APIServer) Shutdown(ctx context.Context) error {
	if a.httpServer == nil {
		logger.Warnf(ctx, "server instance not found, cannot shutdown.")
		return nil
	}

	logger.Infof(ctx, "attempting graceful shutdown of server...")
	err := a.httpServer.Shutdown(ctx)
	if err != nil {
		logger.Errorf(ctx, "server shutdown failed: %v", err)
		return err
	}

	logger.Infof(ctx, "server shutdown completed.")
	return nil
}
