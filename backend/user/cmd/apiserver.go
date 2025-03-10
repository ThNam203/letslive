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

	"go.uber.org/zap"
)

type APIServer struct {
	logger *zap.SugaredLogger
	config config.Config

	errorHandler                 *handlers.ErrorHandler
	healthHandler                *handlers.HealthHandler
	userHandler                  *handlers.UserHandler
	followHandler                *handlers.FollowHandler
	livestreamInformationHandler *handlers.LivestreamInformationHandler

	loggingMiddleware middlewares.Middleware
	corsMiddleware    middlewares.Middleware
}

func NewAPIServer(userHandler *handlers.UserHandler, livestreamInfoHandler *handlers.LivestreamInformationHandler, followHandler *handlers.FollowHandler, cfg config.Config) *APIServer {
	return &APIServer{
		logger: logger.Logger,
		config: cfg,

		errorHandler:                 handlers.NewErrorHandler(),
		healthHandler:                handlers.NewHeathHandler(),
		userHandler:                  userHandler,
		followHandler:                followHandler,
		livestreamInformationHandler: livestreamInfoHandler,

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

	sm.HandleFunc("GET /v1/users", a.userHandler.GetAllUsersHandler) // TODO: remove
	sm.HandleFunc("GET /v1/users/search", a.userHandler.SearchUserHandler)

	sm.HandleFunc("POST /v1/user/{userId}/follow", a.followHandler.FollowHandler)
	sm.HandleFunc("DELETE /v1/user/{userId}/unfollow", a.followHandler.UnfollowHandler)
	sm.HandleFunc("GET /v1/user/{userId}", a.userHandler.GetUserByIdHandler)
	sm.HandleFunc("POST /v1/user", a.userHandler.CreateUserHandler)                 // internal
	sm.HandleFunc("PUT /v1/user/{userId}", a.userHandler.UpdateUserInternalHandler) // internal

	sm.HandleFunc("GET /v1/verify-stream-key", a.userHandler.GetUserByStreamAPIKeyHandler) // internal

	sm.HandleFunc("GET /v1/user/me", a.userHandler.GetCurrentUserHandler)
	sm.HandleFunc("PUT /v1/user/me", a.userHandler.UpdateCurrentUserHandler)
	sm.HandleFunc("PATCH /v1/user/{userId}/set-verified", a.userHandler.SetUserVerifiedHandler) // internal
	sm.HandleFunc("PATCH /v1/user/me/livestream-information", a.livestreamInformationHandler.Update)
	sm.HandleFunc("PATCH /v1/user/me/api-key", a.userHandler.GenerateNewAPIStreamKeyHandler)
	sm.HandleFunc("PATCH /v1/user/me/profile-picture", a.userHandler.UpdateUserProfilePictureHandler)
	sm.HandleFunc("PATCH /v1/user/me/background-picture", a.userHandler.UpdateUserBackgroundPictureHandler)

	sm.HandleFunc("GET /v1/health", a.healthHandler.GetHealthyState)

	sm.HandleFunc("GET /", a.errorHandler.RouteNotFoundHandler)

	finalHandler := a.corsMiddleware.GetMiddleware(sm)
	finalHandler = a.loggingMiddleware.GetMiddleware(finalHandler)

	return finalHandler
}
