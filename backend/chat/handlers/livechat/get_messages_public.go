package livechat

import (
	"net/http"

	"sen1or/letslive/chat/domains"
	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/jsutil"
	"sen1or/letslive/chat/response"
)

func (h *LiveChatHandler) GetMessagesPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	roomID := r.URL.Query().Get("roomId")
	if roomID == "" || jsutil.Length(roomID) > domains.MaxRoomIDLength {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_ROOM_NOT_FOUND, nil, nil, nil))
		return
	}

	messages, err := h.liveChatService.History(ctx, roomID)
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	data := dto.FromLiveMessages(messages)
	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &data, nil, nil))
}
