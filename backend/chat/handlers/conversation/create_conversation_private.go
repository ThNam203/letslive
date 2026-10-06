package conversation

import (
	"encoding/json"
	"net/http"

	"sen1or/letslive/chat/domains"
	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/jsutil"
	"sen1or/letslive/chat/response"
	"sen1or/letslive/chat/services"
)

type createConversationBody struct {
	Type           json.RawMessage `json:"type"`
	ParticipantIDs json.RawMessage `json:"participantIds"`
	Name           json.RawMessage `json:"name"`
}

func (h *ConversationHandler) CreateConversationPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, errResp := utils.GetUserIDFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	var body createConversationBody
	if err := utils.DecodeBody(r, &body); err != nil {
		h.WriteResponse(w, ctx, utils.InvalidPayload())
		return
	}

	conversationType, _ := utils.AsString(body.Type)
	if conversationType != string(domains.ConversationTypeDM) && conversationType != string(domains.ConversationTypeGroup) {
		h.WriteResponse(w, ctx, utils.InvalidInput())
		return
	}

	participantIDs, ok := utils.AsStringArray(body.ParticipantIDs)
	if !ok || len(participantIDs) == 0 {
		h.WriteResponse(w, ctx, utils.InvalidInput())
		return
	}
	for _, id := range participantIDs {
		if jsutil.Length(id) > domains.MaxUserIDLength {
			h.WriteResponse(w, ctx, utils.InvalidInput())
			return
		}
	}

	conversation, created, err := h.conversationService.Create(ctx, services.CreateConversationInput{
		Type:           domains.ConversationType(conversationType),
		CreatorID:      userID,
		ParticipantIDs: participantIDs,
		Name:           utils.OptionalString(body.Name),
	})
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	template := response.RES_SUCC_OK
	if created {
		template = response.RES_SUCC_CREATED
	}
	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(template, dto.FromConversation(conversation), nil, nil))
}
