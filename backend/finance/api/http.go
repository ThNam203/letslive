package api

import (
	"context"
	"fmt"
	"net/http"
	"sen1or/letslive/finance/config"
	"sen1or/letslive/finance/handlers/general"
	"sen1or/letslive/finance/middlewares"
	"sen1or/letslive/finance/pkg/logger"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type APIServer struct {
	httpServer     *http.Server
	logger         *zap.SugaredLogger
	config         *config.Config
	generalHandler *general.GeneralHandler
}

func NewAPIServer(cfg *config.Config) *APIServer {
	return &APIServer{
		logger:         logger.Logger,
		config:         cfg,
		generalHandler: general.NewGeneralHandler(),
	}
}

func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()
	wrap := func(pattern string, fn func(http.ResponseWriter, *http.Request)) {
		sm.Handle(pattern, otelhttp.WithRouteTag(pattern, http.HandlerFunc(fn)))
	}
	wrap("GET /v1/health", a.generalHandler.RouteServiceHealth)
	wrap("GET /", a.generalHandler.RouteNotFoundHandler)

	finalHandler := otelhttp.NewHandler(sm, "/", otelhttp.WithFilter(func(r *http.Request) bool {
		return r.URL.Path != "/v1/health"
	}))
	finalHandler = middlewares.RequestIDMiddleware(finalHandler)
	finalHandler = middlewares.LoggingMiddleware(finalHandler)
	return finalHandler
}

func (a *APIServer) ListenAndServe(ctx context.Context, useTLS bool) error {
	addr := fmt.Sprintf("%s:%d", a.config.Service.APIBindAddress, a.config.Service.APIPort)
	a.httpServer = &http.Server{
		Addr:         addr,
		Handler:      a.getHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	var err error
	if useTLS {
		err = fmt.Errorf("TLS not implemented")
	} else {
		err = a.httpServer.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		logger.Errorf(ctx, "server listener error: %v", err)
		return err
	}
	return nil
}

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
