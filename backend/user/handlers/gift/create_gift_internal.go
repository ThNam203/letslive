package gifthandler

import (
	"context"
	"encoding/json"
	"net/http"

	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/response"
	"sen1or/letslive/shared/pkg/tracer"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
)

func (h *GiftHandler) CreateGiftInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var req dto.CreateGiftInternalRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseWithValidationErrors[any](nil, nil, err))
		return
	}

	senderID, _ := uuid.FromString(req.SenderId)
	recipientID, _ := uuid.FromString(req.RecipientId)
	shopItemID, _ := uuid.FromString(req.ShopItemId)

	ctx, span := tracer.MyTracer.Start(ctx, "gift_handler.create_gift_internal")
	gift, serviceErr := h.giftService.CreateFromPurchase(ctx, senderID, recipientID, shopItemID, req.Quantity, req.Message)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	type giftCreatedResponse struct {
		GiftId string `json:"giftId"`
	}
	data := giftCreatedResponse{GiftId: gift.Id.String()}
	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &data, nil, nil))
}
