package conversation

import (
	"encoding/json"
	"net/http"

	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/response"
	"sen1or/letslive/chat/services"
)

type updateConversationBody struct {
	Name      json.RawMessage `json:"name"`
	AvatarURL json.RawMessage `json:"avatarUrl"`
}

func (h *ConversationHandler) UpdateConversationPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	var body updateConversationBody
	if err := utils.DecodeBody(r, &body); err != nil {
		h.WriteResponse(w, ctx, utils.InvalidPayload())
		return
	}

	conversation, err := h.conversationService.Update(ctx, r.PathValue("id"), userID, services.UpdateConversationInput{
		Name:      utils.OptionalString(body.Name),
		AvatarURL: utils.OptionalString(body.AvatarURL),
	})
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, dto.FromConversation(conversation), nil, nil))
}
