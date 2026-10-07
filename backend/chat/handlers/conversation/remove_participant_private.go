package conversation

import (
	"net/http"

	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/response"
)

func (h *ConversationHandler) RemoveParticipantPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	conversation, err := h.conversationService.RemoveParticipant(ctx, r.PathValue("id"), userID, r.PathValue("userId"))
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, dto.FromConversation(conversation), nil, nil))
}
