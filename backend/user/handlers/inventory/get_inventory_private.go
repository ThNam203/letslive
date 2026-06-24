package inventoryhandler

import (
	"context"
	"net/http"
	"strconv"

	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/response"
	"sen1or/letslive/shared/pkg/tracer"
)

func (h *InventoryHandler) GetInventoryPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	userID, cookieErr := utils.GetUserIdFromCookie(r)
	if cookieErr != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_UNAUTHORIZED, nil, nil, nil))
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	ctx, span := tracer.MyTracer.Start(ctx, "inventory_handler.get_inventory")
	items, total, serviceErr := h.inventoryService.GetByUser(ctx, *userID, page, limit)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	meta := &response.Meta{Page: page, PageSize: limit, Total: total}
	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &items, meta, nil))
}
