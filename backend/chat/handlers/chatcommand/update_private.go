package chatcommand

import (
	"encoding/json"
	"net/http"

	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/response"
	"sen1or/letslive/chat/services"
)

type updateBody struct {
	Name        json.RawMessage `json:"name"`
	Response    json.RawMessage `json:"response"`
	Description json.RawMessage `json:"description"`
}

func (h *ChatCommandHandler) UpdatePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	id, ok := h.commandID(w, r)
	if !ok {
		return
	}

	var body updateBody
	if err := utils.DecodeBody(r, &body); err != nil {
		h.WriteResponse(w, ctx, utils.InvalidPayload())
		return
	}

	command, err := h.chatCommandService.Update(ctx, userID, id, services.UpdateChatCommandInput{
		Name:        utils.OptionalString(body.Name),
		Response:    utils.OptionalString(body.Response),
		Description: utils.OptionalString(body.Description),
	})
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, dto.FromChatCommand(command), nil, nil))
}
