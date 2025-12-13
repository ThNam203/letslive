package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sen1or/letslive/user/config"
	usergrpc "sen1or/letslive/user/handlers/grpc/user"
	"sen1or/letslive/user/handlers/http/follow"
	"sen1or/letslive/user/handlers/http/general"
	"sen1or/letslive/user/handlers/http/livestream_information"
	userhttp "sen1or/letslive/user/handlers/http/user"
	"sen1or/letslive/user/middlewares"
	"sen1or/letslive/user/pkg/logger"

	"time"

	"buf.build/gen/go/letslive/letslive-proto/grpc/go/user/userv1grpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type APIServer struct {
	httpServer *http.Server
	grpcServer *grpc.Server
	logger     *zap.SugaredLogger
	config     *config.Config

	generalHandler               *general.GeneralHandler
	userHandler                  *userhttp.UserHandler
	followHandler                *follow.FollowHandler
	livestreamInformationHandler *livestream_information.LivestreamInformationHandler
	userGRPCHandler              usergrpc.UserGRPCHandler
}

func NewAPIServer(
	userHandler *userhttp.UserHandler,
	livestreamInfoHandler *livestream_information.LivestreamInformationHandler,
	followHandler *follow.FollowHandler,
	userGRPCHandler usergrpc.UserGRPCHandler,
	cfg *config.Config,
) *APIServer {
	return &APIServer{
		logger: logger.Logger,
		config: cfg,

		generalHandler:               general.NewGeneralHandler(),
		userHandler:                  userHandler,
		followHandler:                followHandler,
		livestreamInformationHandler: livestreamInfoHandler,
		userGRPCHandler:              userGRPCHandler,
	}
}

func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()

	wrapHandleFuncWithOtel := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		sm.Handle(pattern, handler)
	}

	wrapHandleFuncWithOtel("POST /v1/upload-file", a.userHandler.UploadSingleFileToMinIOHandler) // TODO: find another way to upload file

	wrapHandleFuncWithOtel("GET /v1/users", a.userHandler.GetAllUsersPublicHandler) // TODO: should change into get random users
	wrapHandleFuncWithOtel("GET /v1/users/search", a.userHandler.SearchUsersPublicHandler)
	wrapHandleFuncWithOtel("GET /v1/user/{userId}", a.userHandler.GetUserByIdPublicHandler)

	wrapHandleFuncWithOtel("POST /v1/user/{userId}/follow", a.followHandler.FollowPrivateHandler)
	wrapHandleFuncWithOtel("DELETE /v1/user/{userId}/unfollow", a.followHandler.UnfollowPrivateHandler)
	wrapHandleFuncWithOtel("GET /v1/user/me", a.userHandler.GetCurrentUserPrivateHandler)
	wrapHandleFuncWithOtel("PUT /v1/user/me", a.userHandler.UpdateCurrentUserPrivateHandler)
	wrapHandleFuncWithOtel("PATCH /v1/user/me/livestream-information", a.livestreamInformationHandler.UpdatePrivateHandler)
	wrapHandleFuncWithOtel("PATCH /v1/user/me/api-key", a.userHandler.GenerateNewAPIStreamKeyPrivateHandler)
	// TODO: change this to not include the FormData
	wrapHandleFuncWithOtel("PATCH /v1/user/me/profile-picture", a.userHandler.UpdateUserProfilePicturePrivateHandler)
	wrapHandleFuncWithOtel("PATCH /v1/user/me/background-picture", a.userHandler.UpdateUserBackgroundPicturePrivateHandler)

	// wrapHandleFuncWithOtel("POST /v1/user", a.userHandler.CreateUserInternalHandler)                        // changed to grpc
	// wrapHandleFuncWithOtel("GET /v1/verify-stream-key", a.userHandler.GetUserByStreamAPIKeyInternalHandler) // changed to grpc
	wrapHandleFuncWithOtel("PUT /v1/user/{userId}", a.userHandler.UpdateUserInternalHandler) // internal

	wrapHandleFuncWithOtel("GET /v1/health", a.generalHandler.RouteServiceHealth)
	wrapHandleFuncWithOtel("GET /", a.generalHandler.RouteNotFoundHandler)

	// TODO: remove filter
	finalHandler := otelhttp.NewHandler(sm, "/", otelhttp.WithFilter(func(r *http.Request) bool {
		return r.URL.Path != "/v1/health" // exclude this path from tracing
	}))
	finalHandler = middlewares.RequestIDMiddleware(finalHandler)
	finalHandler = middlewares.LoggingMiddleware(finalHandler)

	return finalHandler
}

// setupAndStartGRPCServer sets up and starts the gRPC server in a goroutine.
// It returns an error if setup fails, otherwise returns nil and the server runs in the background.
func (a *APIServer) setupAndStartGRPCServer(ctx context.Context) error {
	grpcPort := a.config.Service.APIPort + 1 // Default to API port + 1
	grpcAddr := fmt.Sprintf("%s:%d", a.config.Service.APIBindAddress, grpcPort)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Errorf(ctx, "failed to listen on gRPC address %s: %v", grpcAddr, err)
		return err
	}

	a.grpcServer = grpc.NewServer()
	// Type assert to get the concrete implementation that implements UserServiceServer
	if handler, ok := a.userGRPCHandler.(userv1grpc.UserServiceServer); ok {
		userv1grpc.RegisterUserServiceServer(a.grpcServer, handler)
	} else {
		logger.Errorf(ctx, "gRPC handler does not implement UserServiceServer")
		return fmt.Errorf("gRPC handler does not implement UserServiceServer")
	}

	// Start gRPC server in a goroutine
	go func() {
		logger.Infof(ctx, "starting gRPC server on %s...", grpcAddr)
		if err := a.grpcServer.Serve(lis); err != nil {
			logger.Errorf(ctx, "gRPC server error: %v", err)
		}
	}()

	return nil
}

// setupAndStartHTTPServer sets up and starts the HTTP server.
// It blocks until the server is shut down or an error occurs.
// It returns http.ErrServerClosed on graceful shutdown, otherwise the error.
func (a *APIServer) setupAndStartHTTPServer(ctx context.Context, useTLS bool) error {
	httpAddr := fmt.Sprintf("%s:%d", a.config.Service.APIBindAddress, a.config.Service.APIPort)
	a.httpServer = &http.Server{
		Addr:         httpAddr,
		Handler:      a.getHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start HTTP server (this will block)
	logger.Infof(ctx, "starting HTTP server on %s...", httpAddr)
	var httpErr error
	if useTLS {
		httpErr = fmt.Errorf("TLS not implemented")
	} else {
		httpErr = a.httpServer.ListenAndServe()
	}

	// This line is reached when ListenAndServe returns.
	// It returns http.ErrServerClosed if Shutdown was called gracefully.
	// Otherwise, it returns the error that caused it to stop.
	if httpErr != nil && httpErr != http.ErrServerClosed {
		logger.Errorf(ctx, "HTTP server listener error: %v", httpErr)
		return httpErr
	}

	// If err is nil or http.ErrServerClosed, it means server stopped cleanly or via Shutdown.
	return nil
}

// ListenAndServe sets up and runs the HTTP and gRPC servers.
// it blocks until the servers are shut down or an error occurs.
// it returns http.ErrServerClosed on graceful shutdown, otherwise the error.
func (a *APIServer) ListenAndServe(ctx context.Context, useTLS bool) error {
	// Setup and start gRPC server
	if err := a.setupAndStartGRPCServer(ctx); err != nil {
		return err
	}

	// Setup and start HTTP server (this will block)
	return a.setupAndStartHTTPServer(ctx, useTLS)
}

// shutdown gracefully shuts down the servers without interrupting active connections.
func (a *APIServer) Shutdown(ctx context.Context) error {
	if a.httpServer == nil && a.grpcServer == nil {
		logger.Warnf(ctx, "server instances not found, cannot shutdown.")
		return nil
	}

	logger.Infof(ctx, "attempting graceful shutdown of servers...")

	// Shutdown gRPC server
	if a.grpcServer != nil {
		logger.Infof(ctx, "shutting down gRPC server...")
		a.grpcServer.GracefulStop()
		logger.Infof(ctx, "gRPC server shutdown completed.")
	}

	// Shutdown HTTP server
	if a.httpServer != nil {
		logger.Infof(ctx, "shutting down HTTP server...")
		err := a.httpServer.Shutdown(ctx)
		if err != nil {
			logger.Errorf(ctx, "HTTP server shutdown failed: %v", err)
			return err
		}
		logger.Infof(ctx, "HTTP server shutdown completed.")
	}

	logger.Infof(ctx, "all servers shutdown completed.")
	return nil
}
