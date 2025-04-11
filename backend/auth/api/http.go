package api

import (
	"context"
	"fmt"
	"net/http"
	"sen1or/letslive/auth/config"
	"sen1or/letslive/auth/pkg/discovery"
	"sen1or/letslive/auth/pkg/logger"

	"sen1or/letslive/auth/handlers"
	"sen1or/letslive/auth/middlewares"
	"time"

	"go.uber.org/zap"
)

type APIServer struct {
	httpServer *http.Server
	logger     *zap.SugaredLogger
	config     *config.Config

	authHandler     *handlers.AuthHandler
	responseHandler *handlers.ResponseHandler
	healthHandler   *handlers.HealthHandler
}

func NewAPIServer(
	authHandler *handlers.AuthHandler,
	registry discovery.Registry,
	cfg *config.Config,
) *APIServer {
	return &APIServer{
		logger: logger.Logger,
		config: cfg,

		authHandler:     authHandler,
		responseHandler: handlers.NewResponseHandler(),
		healthHandler:   handlers.NewHeathHandler(),
	}
}

func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()

	sm.HandleFunc("POST /v1/auth/signup", a.authHandler.VerifyOTPAndSignUpHandler)
	sm.HandleFunc("POST /v1/auth/login", a.authHandler.LogInHandler)
	sm.HandleFunc("POST /v1/auth/refresh-token", a.authHandler.RefreshTokenHandler)
	sm.HandleFunc("PATCH /v1/auth/password", a.authHandler.UpdatePasswordHandler)
	sm.HandleFunc("DELETE /v1/auth/logout", a.authHandler.LogOutHandler)
	sm.HandleFunc("POST /v1/auth/verify-email", a.authHandler.RequestEmailVerificationHandler)

	sm.HandleFunc("GET /v1/auth/google", a.authHandler.OAuthGoogleLoginHandler)
	sm.HandleFunc("GET /v1/auth/google/callback", a.authHandler.OAuthGoogleCallBackHandler)

	sm.HandleFunc("GET /v1/health", a.healthHandler.GetHealthyState)
	sm.HandleFunc("GET /", a.responseHandler.RouteNotFoundHandler)

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
