package dmmessage

import (
	"net/http"

	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/jsutil"
	"sen1or/letslive/chat/response"
)

func (h *DmMessageHandler) GetMessagesPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	query := r.URL.Query()
	limit := min(100, max(1, jsutil.IntOr(query.Get("limit"), 50)))

	messages, err := h.dmMessageService.List(ctx, r.PathValue("id"), userID, query.Get("before"), limit)
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	data := dto.FromDmMessages(messages)
	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &data, nil, nil))
}
