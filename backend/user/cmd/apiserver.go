package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sen1or/letslive/user/config"
	"sen1or/letslive/user/handlers"
	"sen1or/letslive/user/middlewares"
	"sen1or/letslive/user/pkg/logger"

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

	go logger.Panicf("server ends: ", server.ListenAndServe())

	logger.Infof("server running on addr: %s", server.Addr)
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

	sm.HandleFunc("POST /v1/upload-file", a.userHandler.UploadSingleFileToMinIOHandler) // TODO: find another way to upload file

	sm.HandleFunc("GET /v1/users", a.userHandler.GetAllUsersPublicHandler) // TODO: should change into get random users
	sm.HandleFunc("GET /v1/users/search", a.userHandler.SearchUsersPublicHandler)
	sm.HandleFunc("GET /v1/user/{userId}", a.userHandler.GetUserByIdPublicHandler)

	sm.HandleFunc("POST /v1/user/{userId}/follow", a.followHandler.FollowPrivateHandler)
	sm.HandleFunc("DELETE /v1/user/{userId}/unfollow", a.followHandler.UnfollowPrivateHandler)
	sm.HandleFunc("GET /v1/user/me", a.userHandler.GetCurrentUserPrivateHandler)
	sm.HandleFunc("PUT /v1/user/me", a.userHandler.UpdateCurrentUserPrivateHandler)
	sm.HandleFunc("PATCH /v1/user/me/livestream-information", a.livestreamInformationHandler.UpdatePrivateHandler)
	sm.HandleFunc("PATCH /v1/user/me/api-key", a.userHandler.GenerateNewAPIStreamKeyPrivateHandler)
	// TODO: change this to not include the FormData
	sm.HandleFunc("PATCH /v1/user/me/profile-picture", a.userHandler.UpdateUserProfilePicturePrivateHandler)
	sm.HandleFunc("PATCH /v1/user/me/background-picture", a.userHandler.UpdateUserBackgroundPicturePrivateHandler)

	sm.HandleFunc("POST /v1/user", a.userHandler.CreateUserInternalHandler)                             // internal
	sm.HandleFunc("PUT /v1/user/{userId}", a.userHandler.UpdateUserInternalHandler)                     // internal
	sm.HandleFunc("GET /v1/verify-stream-key", a.userHandler.GetUserByStreamAPIKeyInternalHandler)      // internal
	sm.HandleFunc("PATCH /v1/user/{userId}/set-verified", a.userHandler.SetUserVerifiedInternalHandler) // internal

	sm.HandleFunc("GET /v1/health", a.healthHandler.GetHealthyStateHandler)
	sm.HandleFunc("GET /", a.errorHandler.RouteNotFoundHandler)

	finalHandler := middlewares.LoggingMiddleware(sm)

	return finalHandler
}
