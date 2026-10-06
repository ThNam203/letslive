package livechat

import (
	"sen1or/letslive/chat/handlers/basehandler"
	"sen1or/letslive/chat/services"
)

type LiveChatHandler struct {
	basehandler.BaseHandler
	liveChatService *services.LiveChatService
}

func NewLiveChatHandler(liveChatService *services.LiveChatService) *LiveChatHandler {
	return &LiveChatHandler{liveChatService: liveChatService}
}
