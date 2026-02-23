package api

import (
	"context"
	"fmt"
	"net/http"
	"sen1or/letslive/livestream/config"
	"sen1or/letslive/livestream/handlers/general"
	"sen1or/letslive/livestream/handlers/livestream"
	"sen1or/letslive/livestream/handlers/vod"
	vodcomment "sen1or/letslive/livestream/handlers/vod_comment"
	"sen1or/letslive/livestream/middlewares"
	"sen1or/letslive/livestream/pkg/logger"

	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type APIServer struct {
	httpServer *http.Server
	logger     *zap.SugaredLogger
	config     *config.Config

	generalHandler    *general.GeneralHandler
	livestreamHandler *livestream.LivestreamHandler
	vodHandler        *vod.VODHandler
	vodCommentHandler *vodcomment.VODCommentHandler
}

func NewAPIServer(livestreamHandler *livestream.LivestreamHandler, vodHandler *vod.VODHandler, vodCommentHandler *vodcomment.VODCommentHandler, cfg *config.Config) *APIServer {
	return &APIServer{
		logger: logger.Logger,
		config: cfg,

		generalHandler:    general.NewGeneralHandler(),
		livestreamHandler: livestreamHandler,
		vodHandler:        vodHandler,
		vodCommentHandler: vodCommentHandler,
	}
}

func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()

	wrap := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		sm.Handle(pattern, http.HandlerFunc(handlerFunc))
	}

	//TODO: change to query livestreams
	wrap("GET /v1/vods", a.vodHandler.GetVODsOfUserPublicHandler)
	wrap("GET /v1/vods/author", a.vodHandler.GetVODsOfAuthorPrivateHandler)
	wrap("GET /v1/vods/{vodId}", a.vodHandler.GetVODByIdPublicHandler)
	wrap("GET /v1/popular-livestreams", a.livestreamHandler.GetRecommendedLivestreamsPublicHandler)
	wrap("GET /v1/livestreams", a.livestreamHandler.GetLivestreamOfUserPublicHandler)
	wrap("GET /v1/popular-vods", a.vodHandler.GetRecommendedVODsPublicHandler)

	wrap("GET /v1/vods/{vodId}/comments", a.vodCommentHandler.GetCommentsPublicHandler)
	wrap("POST /v1/vods/{vodId}/comments", a.vodCommentHandler.CreateCommentPrivateHandler)

	wrap("PATCH /v1/vods/{vodId}", a.vodHandler.UpdateVODMetadataPrivateHandler)
	wrap("DELETE /v1/vods/{vodId}", a.vodHandler.DeleteVODPrivateHandler)

	wrap("GET /v1/vod-comments/{commentId}/replies", a.vodCommentHandler.GetRepliesPublicHandler)
	wrap("DELETE /v1/vod-comments/{commentId}", a.vodCommentHandler.DeleteCommentPrivateHandler)
	wrap("POST /v1/vod-comments/{commentId}/like", a.vodCommentHandler.LikeCommentPrivateHandler)
	wrap("DELETE /v1/vod-comments/{commentId}/like", a.vodCommentHandler.UnlikeCommentPrivateHandler)
	wrap("POST /v1/vod-comments/liked-ids", a.vodCommentHandler.GetUserLikedCommentIdsPrivateHandler)

	wrap("POST /v1/internal/livestreams/{livestreamId}/end", a.livestreamHandler.EndLivestreamAndCreateVODInternalHandler)
	wrap("POST /v1/internal/livestreams", a.livestreamHandler.CreateLivestreamInternalHandler)

	wrap("GET /v1/health", a.generalHandler.RouteServiceHealth)

	wrap("GET /", a.generalHandler.RouteNotFoundHandler)

	// TODO: remove filter
	finalHandler := otelhttp.NewHandler(sm, "/", otelhttp.WithFilter(func(r *http.Request) bool {
		return r.URL.Path != "/v1/health" // exclude this path from tracing
	}))
	finalHandler = middlewares.LoggingMiddleware(finalHandler)
	finalHandler = middlewares.RequestIDMiddleware(finalHandler)

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
