package dmmessage

import (
	"encoding/json"
	"net/http"

	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/response"
)

type editMessageBody struct {
	Text json.RawMessage `json:"text"`
}

func (h *DmMessageHandler) EditMessagePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	var body editMessageBody
	if err := utils.DecodeBody(r, &body); err != nil {
		h.WriteResponse(w, ctx, utils.InvalidPayload())
		return
	}

	text, ok := utils.AsString(body.Text)
	if !ok || text == "" {
		h.WriteResponse(w, ctx, utils.InvalidInput())
		return
	}

	message, err := h.dmMessageService.Edit(ctx, r.PathValue("id"), r.PathValue("msgId"), userID, text)
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, dto.FromDmMessage(message), nil, nil))
}
