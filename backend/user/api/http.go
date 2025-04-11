package api

import (
	"context"
	"fmt"
	"net/http"
	"sen1or/letslive/user/config"
	"sen1or/letslive/user/handlers"
	"sen1or/letslive/user/middlewares"
	"sen1or/letslive/user/pkg/logger"

	"time"

	"go.uber.org/zap"
)

type APIServer struct {
	httpServer *http.Server
	logger     *zap.SugaredLogger
	config     *config.Config

	errorHandler                 *handlers.ErrorHandler
	healthHandler                *handlers.HealthHandler
	userHandler                  *handlers.UserHandler
	followHandler                *handlers.FollowHandler
	livestreamInformationHandler *handlers.LivestreamInformationHandler
}

func NewAPIServer(userHandler *handlers.UserHandler, livestreamInfoHandler *handlers.LivestreamInformationHandler, followHandler *handlers.FollowHandler, cfg *config.Config) *APIServer {
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

	sm.HandleFunc("POST /v1/user", a.userHandler.CreateUserInternalHandler)                        // internal
	sm.HandleFunc("PUT /v1/user/{userId}", a.userHandler.UpdateUserInternalHandler)                // internal
	sm.HandleFunc("GET /v1/verify-stream-key", a.userHandler.GetUserByStreamAPIKeyInternalHandler) // internal

	sm.HandleFunc("GET /v1/health", a.healthHandler.GetHealthyStateHandler)
	sm.HandleFunc("GET /", a.errorHandler.RouteNotFoundHandler)

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
