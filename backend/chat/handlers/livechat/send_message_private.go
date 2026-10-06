package livechat

import (
	"encoding/json"
	"net/http"

	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/response"
)

type sendMessageBody struct {
	RoomID json.RawMessage `json:"roomId"`
	Text   json.RawMessage `json:"text"`
}

// SendMessagePrivateHandler replaces the "message" frame of the old /ws
// socket. The line reaches viewers as a chat.message push on the room topic.
func (h *LiveChatHandler) SendMessagePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	var body sendMessageBody
	if err := utils.DecodeBody(r, &body); err != nil {
		h.WriteResponse(w, ctx, utils.InvalidPayload())
		return
	}

	roomID, _ := utils.AsString(body.RoomID)
	text, ok := utils.AsString(body.Text)
	if !ok {
		h.WriteResponse(w, ctx, utils.InvalidInput())
		return
	}

	message, err := h.liveChatService.Send(ctx, userID, roomID, text)
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_CREATED, dto.FromLiveMessage(message), nil, nil))
}
