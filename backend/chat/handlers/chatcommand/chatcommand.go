package chatcommand

import (
	"net/http"

	"sen1or/letslive/chat/domains"
	"sen1or/letslive/chat/handlers/basehandler"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/jsutil"
	"sen1or/letslive/chat/services"
)

type ChatCommandHandler struct {
	basehandler.BaseHandler
	chatCommandService *services.ChatCommandService
}

func NewChatCommandHandler(chatCommandService *services.ChatCommandService) *ChatCommandHandler {
	return &ChatCommandHandler{chatCommandService: chatCommandService}
}

// commandID returns the {id} path value, or writes invalid_input when it is
// empty or longer than an id can be.
func (h *ChatCommandHandler) commandID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if id == "" || jsutil.Length(id) > domains.MaxChatCommandIDLength {
		h.WriteResponse(w, r.Context(), utils.InvalidInput())
		return "", false
	}
	return id, true
}
