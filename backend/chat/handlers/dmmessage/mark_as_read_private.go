package dmmessage

import (
	"encoding/json"
	"net/http"

	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/response"
)

type markAsReadBody struct {
	MessageID json.RawMessage `json:"messageId"`
}

func (h *DmMessageHandler) MarkAsReadPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	var body markAsReadBody
	if err := utils.DecodeBody(r, &body); err != nil {
		h.WriteResponse(w, ctx, utils.InvalidPayload())
		return
	}

	var messageID *string
	if id, ok := utils.AsString(body.MessageID); ok {
		messageID = &id
	}

	if err := h.dmMessageService.MarkAsRead(ctx, r.PathValue("id"), userID, messageID); err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}
