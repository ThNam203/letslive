package conversation

import (
	"encoding/json"
	"net/http"

	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/response"
)

type addParticipantBody struct {
	UserID   json.RawMessage `json:"userId"`
	Username json.RawMessage `json:"username"`
}

func (h *ConversationHandler) AddParticipantPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	var body addParticipantBody
	if err := utils.DecodeBody(r, &body); err != nil {
		h.WriteResponse(w, ctx, utils.InvalidPayload())
		return
	}

	// username is still required, as in Node, but the stored name always
	// comes from the user service
	targetID, ok := utils.AsString(body.UserID)
	if !ok || targetID == "" || !utils.IsTruthy(body.Username) {
		h.WriteResponse(w, ctx, utils.InvalidInput())
		return
	}

	conversation, err := h.conversationService.AddParticipant(ctx, r.PathValue("id"), userID, targetID)
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, dto.FromConversation(conversation), nil, nil))
}
