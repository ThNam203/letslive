package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sen1or/lets-live/auth/config"
	"sen1or/lets-live/auth/controllers"
	usergateway "sen1or/lets-live/auth/gateway/user/http"
	"sen1or/lets-live/pkg/discovery"
	"sen1or/lets-live/pkg/logger"

	// TODO: add swagger _ "sen1or/lets-live/auth/docs"
	"sen1or/lets-live/auth/handlers"
	"sen1or/lets-live/auth/middlewares"
	"sen1or/lets-live/auth/repositories"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type APIServer struct {
	logger *zap.Logger
	dbConn *pgxpool.Pool
	config config.Config

	authHandler   *handlers.AuthHandler
	errorHandler  *handlers.ErrorHandler
	healthHandler *handlers.HealthHandler

	loggingMiddleware middlewares.Middleware
	corsMiddleware    middlewares.Middleware
}

// TODO: make tls usable
func NewAPIServer(dbConn *pgxpool.Pool, registry discovery.Registry, cfg config.Config) *APIServer {
	var userRepo = repositories.NewAuthRepository(dbConn)
	var refreshTokenRepo = repositories.NewRefreshTokenRepository(dbConn)
	var verifyTokenRepo = repositories.NewVerifyTokenRepo(dbConn)

	var authCtrl = controllers.NewAuthController(userRepo)
	var tokenCtrl = controllers.NewTokenController(refreshTokenRepo, controllers.TokenControllerConfig(cfg.Tokens))
	var verifyTokenCtrl = controllers.NewVerifyTokenController(verifyTokenRepo)

	authServerURL := fmt.Sprintf("http://%s:%d", cfg.Service.Hostname, cfg.Service.APIPort)
	userGateway := usergateway.NewUserGateway(registry)
	var authHandler = handlers.NewAuthHandler(tokenCtrl, authCtrl, verifyTokenCtrl, authServerURL, userGateway)

	var logger, _ = zap.NewProduction()

	return &APIServer{
		logger: logger,
		dbConn: dbConn,
		config: cfg,

		authHandler:   authHandler,
		errorHandler:  handlers.NewErrorHandler(),
		healthHandler: handlers.NewHeathHandler(),

		loggingMiddleware: middlewares.NewLoggingMiddleware(logger),
		corsMiddleware:    middlewares.NewCORSMiddleware(),
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
		if useTLS {
			if _, err := os.Stat(a.config.SSL.ServerCrtFile); err != nil {
				log.Panic("error cant get server cert file", err.Error())
			}

			if _, err := os.Stat(a.config.SSL.ServerKeyFile); err != nil {
				log.Panic("error cant get server key file", err.Error())
			}

			log.Panic("server ends: ", server.ListenAndServeTLS(a.config.SSL.ServerCrtFile, a.config.SSL.ServerKeyFile))
		} else {
			log.Panic("server ends: ", server.ListenAndServe())
		}
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

	sm.HandleFunc("POST /v1/auth/signup", a.authHandler.SignUpHandler)
	sm.HandleFunc("POST /v1/auth/login", a.authHandler.LogInHandler)
	sm.HandleFunc("POST /v1/auth/refresh-token", a.authHandler.RefreshTokenHandler)

	sm.HandleFunc("GET /v1/auth/google", a.authHandler.OAuthGoogleLogin)
	sm.HandleFunc("GET /v1/auth/google/callback", a.authHandler.OAuthGoogleCallBack)
	sm.HandleFunc("GET /v1/auth/email-verify", a.authHandler.VerifyEmailHandler)

	sm.HandleFunc("GET /v1/auth/health", a.healthHandler.GetHealthyState)

	sm.HandleFunc("GET /v1/swagger", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("localhost:%d/swagger/doc.json", a.config.Service.APIPort)),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	sm.HandleFunc("GET /", a.errorHandler.RouteNotFoundHandler)

	finalHandler := a.corsMiddleware.GetMiddleware(sm)
	finalHandler = a.loggingMiddleware.GetMiddleware(finalHandler)

	return finalHandler
}
