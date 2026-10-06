package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"sen1or/letslive/admin/config"
	"sen1or/letslive/admin/handlers/auth"
	"sen1or/letslive/admin/handlers/general"
	"sen1or/letslive/admin/handlers/middleware"
	"sen1or/letslive/shared/middlewares"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type APIServer struct {
	httpServer *http.Server
	logger     *zap.SugaredLogger
	config     *config.Config

	generalHandler *general.GeneralHandler
	authHandler    *auth.AuthHandler
}

func NewAPIServer(authHandler *auth.AuthHandler, cfg *config.Config, db *pgxpool.Pool) *APIServer {
	return &APIServer{
		logger:         logger.Logger,
		config:         cfg,
		generalHandler: general.NewGeneralHandler(db),
		authHandler:    authHandler,
	}
}

func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()

	wrap := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		sm.Handle(pattern, http.HandlerFunc(handlerFunc))
	}

	// Public routes
	wrap("POST /v1/admin/login", a.authHandler.LoginPublicHandler)

	// Private routes (verified in-service — see handlers/middleware/require_admin_auth.go)
	wrap("GET /v1/admin/me", middleware.RequireAdminAuth(a.authHandler.GetMePrivateHandler))

	// Health check
	wrap("GET /v1/health", a.generalHandler.RouteServiceHealth)
	wrap("GET /", a.generalHandler.RouteNotFoundHandler)

	finalHandler := otelhttp.NewHandler(sm, "/", otelhttp.WithFilter(func(r *http.Request) bool {
		return r.URL.Path != "/v1/health"
	}))
	finalHandler = middlewares.LoggingMiddleware(finalHandler)
	finalHandler = middlewares.RequestIDMiddleware(finalHandler)

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
	if err := a.httpServer.Shutdown(ctx); err != nil {
		logger.Errorf(ctx, "server shutdown failed: %v", err)
		return err
	}
	return nil
}
