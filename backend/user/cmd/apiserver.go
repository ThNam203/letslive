package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sen1or/lets-live/pkg/logger"
	"sen1or/lets-live/user/config"
	"sen1or/lets-live/user/handlers"
	"sen1or/lets-live/user/middlewares"

	"time"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type APIServer struct {
	logger *zap.SugaredLogger
	config config.Config

	errorHandler  *handlers.ErrorHandler
	healthHandler *handlers.HealthHandler
	userHandler   *handlers.UserHandler

	loggingMiddleware middlewares.Middleware
	corsMiddleware    middlewares.Middleware
}

// TODO: make tls usable
func NewAPIServer(userHandler *handlers.UserHandler, cfg config.Config) *APIServer {
	return &APIServer{
		logger: logger.Logger,
		config: cfg,

		errorHandler:  handlers.NewErrorHandler(),
		healthHandler: handlers.NewHeathHandler(),
		userHandler:   userHandler,

		loggingMiddleware: middlewares.NewLoggingMiddleware(logger.Logger),
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

	sm.HandleFunc("GET /v1/user/{id}", a.userHandler.GetUserByID)
	sm.HandleFunc("GET /v1/user", a.userHandler.GetUserByQueries)
	sm.HandleFunc("POST /v1/user", a.userHandler.CreateUser)
	sm.HandleFunc("PUT /v1/user/{id}", a.userHandler.UpdateUser)
	sm.HandleFunc("GET /v1/user/me", a.userHandler.GetCurrentUserInfo)

	sm.HandleFunc("GET /v1/user/health", a.healthHandler.GetHealthyState)

	sm.HandleFunc("GET /v1/swagger", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s:%d/swagger/doc.json", a.config.Service.Hostname, a.config.Service.APIPort)),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	sm.HandleFunc("GET /", a.errorHandler.RouteNotFoundHandler)

	finalHandler := a.corsMiddleware.GetMiddleware(sm)
	finalHandler = a.loggingMiddleware.GetMiddleware(finalHandler)

	return finalHandler
}
