package chatcommand

import (
	"encoding/json"
	"net/http"

	"sen1or/letslive/chat/domains"
	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/response"
	"sen1or/letslive/chat/services"
)

type createBody struct {
	Scope       json.RawMessage `json:"scope"`
	Name        json.RawMessage `json:"name"`
	Response    json.RawMessage `json:"response"`
	Description json.RawMessage `json:"description"`
}

func (h *ChatCommandHandler) CreatePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	var body createBody
	if err := utils.DecodeBody(r, &body); err != nil {
		h.WriteResponse(w, ctx, utils.InvalidPayload())
		return
	}

	// a missing or null description defaults to "", as in Node
	description := utils.OptionalString(body.Description)
	commandResponse, responseOK := utils.AsString(body.Response)
	if !responseOK || description.Invalid {
		h.WriteResponse(w, ctx, utils.InvalidInput())
		return
	}
	input := services.CreateChatCommandInput{Response: commandResponse}
	input.Name, _ = utils.AsString(body.Name)
	scope, _ := utils.AsString(body.Scope)
	input.Scope = domains.ChatCommandScope(scope)
	if description.Value != nil {
		input.Description = *description.Value
	}

	command, err := h.chatCommandService.Create(ctx, userID, input)
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_CREATED, dto.FromChatCommand(command), nil, nil))
}
