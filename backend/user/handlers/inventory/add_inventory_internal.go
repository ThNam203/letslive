package inventoryhandler

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

func (h *InventoryHandler) AddInventoryInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var req dto.AddInventoryInternalRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseWithValidationErrors[any](nil, nil, err))
		return
	}

	userID, _ := uuid.FromString(req.UserId)
	shopItemID, _ := uuid.FromString(req.ShopItemId)

	ctx, span := tracer.MyTracer.Start(ctx, "inventory_handler.add_inventory_internal")
	_, serviceErr := h.inventoryService.AddItems(ctx, userID, shopItemID, req.Quantity)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}
