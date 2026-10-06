package conversation

import (
	"net/http"

	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/jsutil"
	"sen1or/letslive/chat/response"
)

func (h *ConversationHandler) GetConversationsPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	query := r.URL.Query()
	page := max(0, jsutil.IntOr(query.Get("page"), 0))
	limit := min(50, max(1, jsutil.IntOr(query.Get("limit"), 20)))

	conversations, total, err := h.conversationService.List(ctx, userID, page, limit)
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	data := dto.FromConversations(conversations)
	meta := &response.Meta{Page: page, PageSize: limit, Total: int(total)}
	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &data, meta, nil))
}
