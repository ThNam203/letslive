package shopitemhandler

import (
	"context"
	"net/http"

	"sen1or/letslive/finance/dto"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/tracer"

	"github.com/google/uuid"
)

func (h *ShopItemHandler) GetItemPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_item_public_handler.shop_item_service.get_by_id")
	item, serviceErr := h.shopItemService.GetById(ctx, id)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	result := dto.ShopItemResponseDTO{
		Id:           item.Id,
		Name:         item.Name,
		Description:  item.Description,
		ImageURL:     item.ImageURL,
		AnimationURL: item.AnimationURL,
		Price:        item.Price,
		CreatedAt:    item.CreatedAt,
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &result, nil, nil))
}
