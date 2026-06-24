package shopitemhandler

import (
	"context"
	"net/http"

	"sen1or/letslive/finance/dto"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/tracer"
)

func (h *ShopItemHandler) GetItemsPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	ctx, span := tracer.MyTracer.Start(ctx, "get_items_public_handler.shop_item_service.list")
	items, serviceErr := h.shopItemService.List(ctx)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	result := make([]dto.ShopItemResponseDTO, len(items))
	for i, item := range items {
		result[i] = dto.ShopItemResponseDTO{
			Id:           item.Id,
			Name:         item.Name,
			Description:  item.Description,
			ImageURL:     item.ImageURL,
			AnimationURL: item.AnimationURL,
			Price:        item.Price,
			CreatedAt:    item.CreatedAt,
		}
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &result, nil, nil))
}
