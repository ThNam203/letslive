package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sen1or/lets-live/livestream/config"
	"sen1or/lets-live/livestream/handlers"
	"sen1or/lets-live/livestream/middlewares"
	"sen1or/lets-live/pkg/logger"

	"time"

	"go.uber.org/zap"
)

type APIServer struct {
	logger *zap.SugaredLogger
	config config.Config

	errorHandler      *handlers.ErrorHandler
	healthHandler     *handlers.HealthHandler
	livestreamHandler *handlers.LivestreamHandler

	loggingMiddleware middlewares.Middleware
	corsMiddleware    middlewares.Middleware
}

func NewAPIServer(livestreamHandler *handlers.LivestreamHandler, cfg config.Config) *APIServer {
	return &APIServer{
		logger: logger.Logger,
		config: cfg,

		errorHandler:      handlers.NewErrorHandler(),
		healthHandler:     handlers.NewHeathHandler(),
		livestreamHandler: livestreamHandler,

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

	go log.Panic("server ends: ", server.ListenAndServe())

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

func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()

	sm.HandleFunc("POST /v1/livestream", a.livestreamHandler.CreateLivestream)
	sm.HandleFunc("PUT /v1/livestream", a.livestreamHandler.UpdateLivestream)
	sm.HandleFunc("GET /v1/livestream/{livestreamId}", a.livestreamHandler.GetLivestreamsById)
	sm.HandleFunc("GET /v1/livestream", a.livestreamHandler.GetLivestreamsOfUser)

	sm.HandleFunc("GET /v1/health", a.healthHandler.GetHealthyState)

	sm.HandleFunc("GET /", a.errorHandler.RouteNotFoundHandler)

	finalHandler := a.corsMiddleware.GetMiddleware(sm)
	finalHandler = a.loggingMiddleware.GetMiddleware(finalHandler)

	return finalHandler
}
