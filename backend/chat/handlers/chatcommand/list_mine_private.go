package chatcommand

import (
	"net/http"

	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/response"
)

func (h *ChatCommandHandler) ListMinePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	user, channel, err := h.chatCommandService.ListMine(ctx, userID)
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	data := &dto.MyChatCommandsResponse{User: dto.FromChatCommands(user), Channel: dto.FromChatCommands(channel)}
	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, data, nil, nil))
}
