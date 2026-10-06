package dmmessage

import (
	"encoding/json"
	"net/http"

	"sen1or/letslive/chat/domains"
	"sen1or/letslive/chat/dto"
	"sen1or/letslive/chat/handlers/utils"
	"sen1or/letslive/chat/response"
	"sen1or/letslive/chat/services"
)

type sendMessageBody struct {
	Text      json.RawMessage `json:"text"`
	Type      json.RawMessage `json:"type"`
	ImageURLs json.RawMessage `json:"imageUrls"`
	ReplyTo   json.RawMessage `json:"replyTo"`
}

func (h *DmMessageHandler) SendMessagePrivateHandler(w http.ResponseWriter, r *http.Request) {
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

	text, ok := utils.AsString(body.Text)
	if !ok || text == "" {
		h.WriteResponse(w, ctx, utils.InvalidInput())
		return
	}

	// a falsy type means the default; any other value must be a known type
	messageType := domains.DmMessageTypeText
	if utils.IsTruthy(body.Type) {
		raw, _ := utils.AsString(body.Type)
		messageType = domains.DmMessageType(raw)
		if !messageType.Valid() {
			h.WriteResponse(w, ctx, utils.InvalidInput())
			return
		}
	}

	var imageURLs []string
	if utils.IsTruthy(body.ImageURLs) {
		if imageURLs, ok = utils.AsStringArray(body.ImageURLs); !ok {
			h.WriteResponse(w, ctx, utils.InvalidInput())
			return
		}
	}

	replyTo, _ := utils.AsString(body.ReplyTo)

	message, participantIDs, err := h.dmMessageService.Send(ctx, services.SendDmMessageInput{
		ConversationID: r.PathValue("id"),
		SenderID:       userID,
		Text:           text,
		Type:           messageType,
		ImageURLs:      imageURLs,
		ReplyTo:        replyTo,
	})
	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	data := &dto.SentDmMessageResponse{DmMessageResponse: *dto.FromDmMessage(message), ParticipantIDs: participantIDs}
	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_CREATED, data, nil, nil))
}
