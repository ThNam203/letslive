package conversation

import (
	"sen1or/letslive/chat/handlers/basehandler"
	"sen1or/letslive/chat/services"
)

type ConversationHandler struct {
	basehandler.BaseHandler
	conversationService *services.ConversationService
}

func NewConversationHandler(conversationService *services.ConversationService) *ConversationHandler {
	return &ConversationHandler{conversationService: conversationService}
}
