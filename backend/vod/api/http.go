package api

import (
	"context"
	"fmt"
	"net/http"
	"sen1or/letslive/vod/config"
	"sen1or/letslive/vod/handlers/general"
	"sen1or/letslive/vod/handlers/vod"
	vodcomment "sen1or/letslive/vod/handlers/vod_comment"
	"sen1or/letslive/shared/middlewares"
	"sen1or/letslive/shared/pkg/logger"

	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type APIServer struct {
	httpServer *http.Server
	logger     *zap.SugaredLogger
	config     *config.Config

	generalHandler    *general.GeneralHandler
	vodHandler        *vod.VODHandler
	vodCommentHandler *vodcomment.VODCommentHandler
}

func NewAPIServer(vodHandler *vod.VODHandler, vodCommentHandler *vodcomment.VODCommentHandler, cfg *config.Config, db *pgxpool.Pool) *APIServer {
	return &APIServer{
		logger: logger.Logger,
		config: cfg,

		generalHandler:    general.NewGeneralHandler(db),
		vodHandler:        vodHandler,
		vodCommentHandler: vodCommentHandler,
	}
}

func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()

	wrap := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		sm.Handle(pattern, http.HandlerFunc(handlerFunc))
	}

	// Public VOD routes
	wrap("GET /v1/vods", a.vodHandler.GetVODsOfUserPublicHandler)
	wrap("GET /v1/vods/{vodId}", a.vodHandler.GetVODByIdPublicHandler)
	wrap("POST /v1/vods/{vodId}/view", a.vodHandler.RegisterViewPublicHandler)
	wrap("GET /v1/popular-vods", a.vodHandler.GetRecommendedVODsPublicHandler)

	// Private VOD routes (require JWT via Kong)
	wrap("GET /v1/vods/author", a.vodHandler.GetVODsOfAuthorPrivateHandler)
	wrap("POST /v1/vods/upload", a.vodHandler.UploadVODPrivateHandler)
	wrap("PATCH /v1/vods/{vodId}", a.vodHandler.UpdateVODMetadataPrivateHandler)
	wrap("DELETE /v1/vods/{vodId}", a.vodHandler.DeleteVODPrivateHandler)

	// Public VOD comment routes
	wrap("GET /v1/vods/{vodId}/comments", a.vodCommentHandler.GetCommentsPublicHandler)
	wrap("GET /v1/vod-comments/{commentId}/replies", a.vodCommentHandler.GetRepliesPublicHandler)

	// Private VOD comment routes
	wrap("POST /v1/vods/{vodId}/comments", a.vodCommentHandler.CreateCommentPrivateHandler)
	wrap("DELETE /v1/vod-comments/{commentId}", a.vodCommentHandler.DeleteCommentPrivateHandler)
	wrap("POST /v1/vod-comments/{commentId}/like", a.vodCommentHandler.LikeCommentPrivateHandler)
	wrap("DELETE /v1/vod-comments/{commentId}/like", a.vodCommentHandler.UnlikeCommentPrivateHandler)
	wrap("POST /v1/vod-comments/liked-ids", a.vodCommentHandler.GetUserLikedCommentIdsPrivateHandler)

	// Internal routes (service-to-service, no JWT)
	wrap("POST /v1/internal/vods", a.vodHandler.CreateVODInternalHandler)
	wrap("PATCH /v1/internal/vods/{vodId}/status", a.vodHandler.UpdateVODStatusInternalHandler)

	// Health check
	wrap("GET /v1/health", a.generalHandler.RouteServiceHealth)
	wrap("GET /", a.generalHandler.RouteNotFoundHandler)

	finalHandler := otelhttp.NewHandler(sm, "/", otelhttp.WithFilter(func(r *http.Request) bool {
		return r.URL.Path != "/v1/health"
	}))
	finalHandler = middlewares.MaxBodySizeMiddleware(1<<20)(finalHandler) // 1MB default; upload handler overrides with its own 2GB limit
	finalHandler = middlewares.LoggingMiddleware(finalHandler)
	finalHandler = middlewares.RequestIDMiddleware(finalHandler)

	return finalHandler
}

func (a *APIServer) ListenAndServe(ctx context.Context, useTLS bool) error {
	addr := fmt.Sprintf("%s:%d", a.config.Service.APIBindAddress, a.config.Service.APIPort)

	a.httpServer = &http.Server{
		Addr:              addr,
		Handler:           a.getHandler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Minute,
		WriteTimeout:      5 * time.Minute,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	var err error
	if useTLS {
		err = fmt.Errorf("TLS not implemented")
	} else {
		err = a.httpServer.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		logger.Errorf(ctx, "server listener error: %v", err)
		return err
	}

	return nil
}

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
