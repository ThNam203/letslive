package dmmessage

import (
	"net/http"

	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/response"
)

func (h *DmMessageHandler) DeleteMessagePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	if err := h.dmMessageService.Delete(ctx, r.PathValue("id"), r.PathValue("msgId"), userID); err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}
