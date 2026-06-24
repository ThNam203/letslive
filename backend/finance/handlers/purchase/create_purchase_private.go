package purchasehandler

import (
	"context"
	"encoding/json"
	"net/http"

	"sen1or/letslive/finance/dto"
	"sen1or/letslive/finance/handlers/utils"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/tracer"

	"github.com/go-playground/validator/v10"
)

func (h *PurchaseHandler) CreatePurchasePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	actorID, errResp := utils.GetUserIdFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	var req dto.PurchaseRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseWithValidationErrors[any](nil, nil, err))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "purchase_handler.create_purchase")
	result, serviceErr := h.purchaseService.Purchase(ctx, *actorID, req)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, result, nil, nil))
}
