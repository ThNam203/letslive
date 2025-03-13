package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sen1or/letslive/livestream/config"
	"sen1or/letslive/livestream/handlers"
	"sen1or/letslive/livestream/middlewares"
	"sen1or/letslive/livestream/pkg/logger"

	"time"

	"go.uber.org/zap"
)

type APIServer struct {
	logger *zap.SugaredLogger
	config config.Config

	errorHandler      *handlers.ErrorHandler
	healthHandler     *handlers.HealthHandler
	livestreamHandler *handlers.LivestreamHandler
}

func NewAPIServer(livestreamHandler *handlers.LivestreamHandler, cfg config.Config) *APIServer {
	return &APIServer{
		logger: logger.Logger,
		config: cfg,

		errorHandler:      handlers.NewErrorHandler(),
		healthHandler:     handlers.NewHeathHandler(),
		livestreamHandler: livestreamHandler,
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
	//TODO: change to query livestreams
	sm.HandleFunc("GET /v1/livestreams", a.livestreamHandler.GetLivestreamsOfUserPublicHandler)
	sm.HandleFunc("GET /v1/livestreams/author", a.livestreamHandler.GetLivestreamsOfUserAuthorHandler)
	sm.HandleFunc("GET /v1/livestreams/{livestreamId}", a.livestreamHandler.GetLivestreamByIdPublicHandler)
	sm.HandleFunc("PATCH /v1/livestreams/{livestreamId}", a.livestreamHandler.UpdateLivestreamHandler)
	sm.HandleFunc("DELETE /v1/livestreams/{livestreamId}", a.livestreamHandler.DeleteLivestreamHandler)

	sm.HandleFunc("GET /v1/livestreamings", a.livestreamHandler.GetLivestreamingsHandler)
	sm.HandleFunc("GET /v1/popular-vods", a.livestreamHandler.GetPopularVODs)
	sm.HandleFunc("GET /v1/is-streaming", a.livestreamHandler.CheckIsUserLivestreamingHandler)

	sm.HandleFunc("PUT /v1/internal/livestreams/{livestreamId}", a.livestreamHandler.UpdateLivestreamInternalHandler)
	sm.HandleFunc("POST /v1/internal/livestreams", a.livestreamHandler.CreateLivestreamInternalHandler)

	sm.HandleFunc("GET /v1/health", a.healthHandler.GetHealthyState)

	sm.HandleFunc("GET /", a.errorHandler.RouteNotFoundHandler)

	finalHandler := middlewares.CORSMiddleware(sm)
	finalHandler = middlewares.LoggingMiddleware(a.logger, finalHandler)

	return finalHandler
}
