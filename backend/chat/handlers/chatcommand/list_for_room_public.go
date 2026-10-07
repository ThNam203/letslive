package chatcommand

import (
	"net/http"

	"sen1or/letslive/chat/domains"
	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/jsutil"
	"sen1or/letslive/chat/response"
)

// ListForRoomPublicHandler is public; a signed-in viewer also gets their own
// user commands.
func (h *ChatCommandHandler) ListForRoomPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	roomID := r.URL.Query().Get("roomId")
	if roomID == "" || jsutil.Length(roomID) > domains.MaxRoomIDLength {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_ROOM_NOT_FOUND, nil, nil, nil))
		return
	}

	viewerID, _ := utils.GetUserIDFromCookie(r)

	commands, err := h.chatCommandService.ListForRoom(ctx, roomID, viewerID)
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	data := dto.FromChatCommands(commands)
	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &data, nil, nil))
}
