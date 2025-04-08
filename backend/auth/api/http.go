package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sen1or/letslive/auth/config"
	"sen1or/letslive/auth/pkg/discovery"
	"sen1or/letslive/auth/pkg/logger"

	"sen1or/letslive/auth/handlers"
	"sen1or/letslive/auth/middlewares"
	"time"

	"go.uber.org/zap"
)

type APIServer struct {
	logger *zap.SugaredLogger
	config config.Config

	authHandler     *handlers.AuthHandler
	responseHandler *handlers.ResponseHandler
	healthHandler   *handlers.HealthHandler
}

func NewAPIServer(
	authHandler *handlers.AuthHandler,
	registry discovery.Registry,
	cfg config.Config,
) *APIServer {
	return &APIServer{
		logger: logger.Logger,
		config: cfg,

		authHandler:     authHandler,
		responseHandler: handlers.NewResponseHandler(),
		healthHandler:   handlers.NewHeathHandler(),
	}
}

func (a *APIServer) ListenAndServe(useTLS bool) {
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.config.Service.APIBindAddress, a.config.Service.APIPort),
		Handler:      a.getHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		log.Panic("server ends: ", server.ListenAndServe())
	}()

	log.Printf("server running on addr: %s", server.Addr)
	<-quit

	// Shutdown gracefully
	logger.Infow("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("server shutdown failed: %+v", err)
	}

	logger.Infow("server exited gracefully")
}

// @title           Let's Live API
// @version         0.1
// @description     The server API

// @contact.name   Nam Huynh
// @contact.email  hthnam203@gmail.com

// @host      localhost:8000
// @BasePath  /v1
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
