package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"sen1or/letslive/chat/config"
	"sen1or/letslive/chat/handlers/chatcommand"
	"sen1or/letslive/chat/handlers/conversation"
	"sen1or/letslive/chat/handlers/dmmessage"
	"sen1or/letslive/chat/handlers/general"
	"sen1or/letslive/chat/handlers/livechat"
	"sen1or/letslive/shared/middlewares"
	"sen1or/letslive/shared/pkg/logger"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Handlers groups the route handlers by feature.
type Handlers struct {
	General      *general.GeneralHandler
	Conversation *conversation.ConversationHandler
	DmMessage    *dmmessage.DmMessageHandler
	LiveChat     *livechat.LiveChatHandler
	ChatCommand  *chatcommand.ChatCommandHandler
}

type APIServer struct {
	config     *config.Config
	httpServer *http.Server
	handlers   Handlers
}

func NewAPIServer(cfg *config.Config, handlers Handlers) *APIServer {
	return &APIServer{config: cfg, handlers: handlers}
}

func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()
	h := a.handlers

	// Public routes
	sm.HandleFunc("GET /v1/messages", h.LiveChat.GetMessagesPublicHandler)
	sm.HandleFunc("GET /v1/chat-commands", h.ChatCommand.ListForRoomPublicHandler)

	// Private routes (JWT checked by the gateway)
	sm.HandleFunc("POST /v1/messages", h.LiveChat.SendMessagePrivateHandler)

	sm.HandleFunc("GET /v1/chat-commands/mine", h.ChatCommand.ListMinePrivateHandler)
	sm.HandleFunc("POST /v1/chat-commands", h.ChatCommand.CreatePrivateHandler)
	sm.HandleFunc("PATCH /v1/chat-commands/{id}", h.ChatCommand.UpdatePrivateHandler)
	sm.HandleFunc("DELETE /v1/chat-commands/{id}", h.ChatCommand.DeletePrivateHandler)

	sm.HandleFunc("GET /v1/conversations/unread-counts", h.Conversation.GetUnreadCountsPrivateHandler)
	sm.HandleFunc("GET /v1/conversations", h.Conversation.GetConversationsPrivateHandler)
	sm.HandleFunc("POST /v1/conversations", h.Conversation.CreateConversationPrivateHandler)
	sm.HandleFunc("GET /v1/conversations/{id}", h.Conversation.GetConversationPrivateHandler)
	sm.HandleFunc("PUT /v1/conversations/{id}", h.Conversation.UpdateConversationPrivateHandler)
	sm.HandleFunc("DELETE /v1/conversations/{id}", h.Conversation.LeaveConversationPrivateHandler)
	sm.HandleFunc("POST /v1/conversations/{id}/participants", h.Conversation.AddParticipantPrivateHandler)
	sm.HandleFunc("DELETE /v1/conversations/{id}/participants/{userId}", h.Conversation.RemoveParticipantPrivateHandler)

	sm.HandleFunc("GET /v1/conversations/{id}/messages", h.DmMessage.GetMessagesPrivateHandler)
	sm.HandleFunc("POST /v1/conversations/{id}/messages", h.DmMessage.SendMessagePrivateHandler)
	sm.HandleFunc("PATCH /v1/conversations/{id}/messages/{msgId}", h.DmMessage.EditMessagePrivateHandler)
	sm.HandleFunc("DELETE /v1/conversations/{id}/messages/{msgId}", h.DmMessage.DeleteMessagePrivateHandler)
	sm.HandleFunc("POST /v1/conversations/{id}/read", h.DmMessage.MarkAsReadPrivateHandler)
	sm.HandleFunc("POST /v1/conversations/{id}/typing", h.DmMessage.SendTypingPrivateHandler)

	sm.HandleFunc("GET /v1/health", h.General.RouteServiceHealth)
	// any method, so an unknown route or method gets the JSON 404 as in Node
	sm.HandleFunc("/", h.General.RouteNotFoundHandler)

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
