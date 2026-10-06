package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"sen1or/letslive/chat/config"
	"sen1or/letslive/chat/handlers/conversation"
	"sen1or/letslive/chat/handlers/dmmessage"
	"sen1or/letslive/chat/handlers/general"
	"sen1or/letslive/shared/middlewares"
	"sen1or/letslive/shared/pkg/logger"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type APIServer struct {
	config     *config.Config
	httpServer *http.Server

	generalHandler      *general.GeneralHandler
	conversationHandler *conversation.ConversationHandler
	dmMessageHandler    *dmmessage.DmMessageHandler
}

func NewAPIServer(
	cfg *config.Config,
	generalHandler *general.GeneralHandler,
	conversationHandler *conversation.ConversationHandler,
	dmMessageHandler *dmmessage.DmMessageHandler,
) *APIServer {
	return &APIServer{
		config:              cfg,
		generalHandler:      generalHandler,
		conversationHandler: conversationHandler,
		dmMessageHandler:    dmMessageHandler,
	}
}

func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()

	// Private routes (JWT checked by Kong)
	sm.HandleFunc("GET /v1/conversations/unread-counts", a.conversationHandler.GetUnreadCountsPrivateHandler)
	sm.HandleFunc("GET /v1/conversations", a.conversationHandler.GetConversationsPrivateHandler)
	sm.HandleFunc("POST /v1/conversations", a.conversationHandler.CreateConversationPrivateHandler)
	sm.HandleFunc("GET /v1/conversations/{id}", a.conversationHandler.GetConversationPrivateHandler)
	sm.HandleFunc("PUT /v1/conversations/{id}", a.conversationHandler.UpdateConversationPrivateHandler)
	sm.HandleFunc("DELETE /v1/conversations/{id}", a.conversationHandler.LeaveConversationPrivateHandler)
	sm.HandleFunc("POST /v1/conversations/{id}/participants", a.conversationHandler.AddParticipantPrivateHandler)
	sm.HandleFunc("DELETE /v1/conversations/{id}/participants/{userId}", a.conversationHandler.RemoveParticipantPrivateHandler)

	sm.HandleFunc("GET /v1/conversations/{id}/messages", a.dmMessageHandler.GetMessagesPrivateHandler)
	sm.HandleFunc("POST /v1/conversations/{id}/messages", a.dmMessageHandler.SendMessagePrivateHandler)
	sm.HandleFunc("PATCH /v1/conversations/{id}/messages/{msgId}", a.dmMessageHandler.EditMessagePrivateHandler)
	sm.HandleFunc("DELETE /v1/conversations/{id}/messages/{msgId}", a.dmMessageHandler.DeleteMessagePrivateHandler)
	sm.HandleFunc("POST /v1/conversations/{id}/read", a.dmMessageHandler.MarkAsReadPrivateHandler)
	sm.HandleFunc("POST /v1/conversations/{id}/typing", a.dmMessageHandler.SendTypingPrivateHandler)

	sm.HandleFunc("GET /v1/health", a.generalHandler.RouteServiceHealth)
	// any method, so an unknown route or method gets the JSON 404 as in Node
	sm.HandleFunc("/", a.generalHandler.RouteNotFoundHandler)

	finalHandler := otelhttp.NewHandler(sm, "/", otelhttp.WithFilter(func(r *http.Request) bool {
		return r.URL.Path != "/v1/health"
	}))
	finalHandler = middlewares.LoggingMiddleware(finalHandler)
	finalHandler = middlewares.RequestIDMiddleware(finalHandler)

	return finalHandler
}

// ListenAndServe blocks until the server stops; a graceful shutdown returns nil.
func (a *APIServer) ListenAndServe(ctx context.Context) error {
	a.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.config.Service.APIBindAddress, a.config.Service.APIPort),
		Handler:      a.getHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err := a.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Errorf(ctx, "server listener error: %v", err)
		return err
	}
	return nil
}

func (a *APIServer) Shutdown(ctx context.Context) error {
	if a.httpServer == nil {
		return nil
	}
	return a.httpServer.Shutdown(ctx)
}
