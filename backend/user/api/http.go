package api

import (
	"context"
	"fmt"
	"net/http"
	"sen1or/letslive/user/config"
	"sen1or/letslive/user/handlers/follow"
	"sen1or/letslive/user/handlers/general"
	"sen1or/letslive/user/handlers/livestream_information"
	"sen1or/letslive/user/handlers/notification"
	"sen1or/letslive/user/handlers/user"
	"sen1or/letslive/user/middlewares"
	"sen1or/letslive/user/pkg/logger"

	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type APIServer struct {
	httpServer *http.Server
	logger     *zap.SugaredLogger
	config     *config.Config

	generalHandler               *general.GeneralHandler
	userHandler                  *user.UserHandler
	followHandler                *follow.FollowHandler
	livestreamInformationHandler *livestream_information.LivestreamInformationHandler
	notificationHandler          *notification.NotificationHandler
}

func NewAPIServer(userHandler *user.UserHandler, livestreamInfoHandler *livestream_information.LivestreamInformationHandler, followHandler *follow.FollowHandler, notificationHandler *notification.NotificationHandler, cfg *config.Config) *APIServer {
	return &APIServer{
		logger: logger.Logger,
		config: cfg,

		generalHandler:               general.NewGeneralHandler(),
		userHandler:                  userHandler,
		followHandler:                followHandler,
		livestreamInformationHandler: livestreamInfoHandler,
		notificationHandler:          notificationHandler,
	}
}

func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()

	wrap := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		sm.Handle(pattern, http.HandlerFunc(handlerFunc))
	}

	wrap("POST /v1/upload-file", a.userHandler.UploadSingleFileToMinIOHandler) // TODO: find another way to upload file

	wrap("GET /v1/users", a.userHandler.GetAllUsersPublicHandler) // TODO: should change into get random users
	wrap("GET /v1/users/search", a.userHandler.SearchUsersPublicHandler)
	wrap("GET /v1/user/{userId}", a.userHandler.GetUserByIdPublicHandler)

	wrap("POST /v1/user/{userId}/follow", a.followHandler.FollowPrivateHandler)
	wrap("DELETE /v1/user/{userId}/unfollow", a.followHandler.UnfollowPrivateHandler)
	wrap("GET /v1/user/me", a.userHandler.GetCurrentUserPrivateHandler)
	wrap("PUT /v1/user/me", a.userHandler.UpdateCurrentUserPrivateHandler)
	wrap("PATCH /v1/user/me/livestream-information", a.livestreamInformationHandler.UpdatePrivateHandler)
	wrap("PATCH /v1/user/me/api-key", a.userHandler.GenerateNewAPIStreamKeyPrivateHandler)
	// TODO: change this to not include the FormData
	wrap("PATCH /v1/user/me/profile-picture", a.userHandler.UpdateUserProfilePicturePrivateHandler)
	wrap("PATCH /v1/user/me/background-picture", a.userHandler.UpdateUserBackgroundPicturePrivateHandler)

	// notifications
	wrap("GET /v1/user/me/notifications", a.notificationHandler.GetNotificationsPrivateHandler)
	wrap("GET /v1/user/me/notifications/unread-count", a.notificationHandler.GetUnreadCountPrivateHandler)
	wrap("PATCH /v1/user/me/notifications/{notificationId}/read", a.notificationHandler.MarkAsReadPrivateHandler)
	wrap("PATCH /v1/user/me/notifications/read-all", a.notificationHandler.MarkAllAsReadPrivateHandler)
	wrap("DELETE /v1/user/me/notifications/{notificationId}", a.notificationHandler.DeleteNotificationPrivateHandler)
	wrap("POST /v1/notifications", a.notificationHandler.CreateNotificationInternalHandler) // internal

	wrap("POST /v1/user", a.userHandler.CreateUserInternalHandler)                        // internal
	wrap("PUT /v1/user/{userId}", a.userHandler.UpdateUserInternalHandler)                // internal
	wrap("GET /v1/verify-stream-key", a.userHandler.GetUserByStreamAPIKeyInternalHandler) // internal

	wrap("GET /v1/health", a.generalHandler.RouteServiceHealth)
	wrap("GET /", a.generalHandler.RouteNotFoundHandler)

	// TODO: remove filter
	finalHandler := otelhttp.NewHandler(sm, "/", otelhttp.WithFilter(func(r *http.Request) bool {
		return r.URL.Path != "/v1/health" // exclude this path from tracing
	}))
	finalHandler = middlewares.RequestIDMiddleware(finalHandler)
	finalHandler = middlewares.LoggingMiddleware(finalHandler)

	return finalHandler
}

// ListenAndServe sets up and runs the HTTP server.
// it blocks until the server is shut down or an error occurs.
// it returns http.ErrServerClosed on graceful shutdown, otherwise the error.
func (a *APIServer) ListenAndServe(ctx context.Context, useTLS bool) error { // Changed signature to return error
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
		logger.Errorf(ctx, "server listener error: %v", err)
		return err
	}

	// If err is nil or http.ErrServerClosed, it means server stopped cleanly or via Shutdown.
	return nil
}

// shutdown gracefully shuts down the server without interrupting active connections.
func (a *APIServer) Shutdown(ctx context.Context) error {
	if a.httpServer == nil {
		logger.Warnf(ctx, "server instance not found, cannot shutdown.")
		return nil
	}

	logger.Infof(ctx, "attempting graceful shutdown of server...")
	err := a.httpServer.Shutdown(ctx)
	if err != nil {
		logger.Errorf(ctx, "server shutdown failed: %v", err)
		return err
	}

	logger.Infof(ctx, "server shutdown completed.")
	return nil
}
