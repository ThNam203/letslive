package dmmessage

import (
	"encoding/json"
	"net/http"

	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/response"
)

type typingBody struct {
	State json.RawMessage `json:"state"`
}

// SendTypingPrivateHandler replaces the dm:typing_start / dm:typing_stop
// frames of the old DM socket. Nothing is stored; the other participants get
// a push.
func (h *DmMessageHandler) SendTypingPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	var body typingBody
	if err := utils.DecodeBody(r, &body); err != nil {
		h.WriteResponse(w, ctx, utils.InvalidPayload())
		return
	}

	state, _ := utils.AsString(body.State)
	if state != "start" && state != "stop" {
		h.WriteResponse(w, ctx, utils.InvalidInput())
		return
	}

	if err := h.dmMessageService.Typing(ctx, r.PathValue("id"), userID, state == "start"); err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}
