package conversation

import (
	"net/http"

	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/response"
)

func (h *ConversationHandler) GetUnreadCountsPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	counts, err := h.conversationService.UnreadCounts(ctx, userID)
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &counts, nil, nil))
}
