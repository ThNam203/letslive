package gifthandler

import (
	"context"
	"net/http"
	"strconv"

	"sen1or/letslive/user/response"
	"sen1or/letslive/shared/pkg/tracer"

	"github.com/gofrs/uuid/v5"
)

func (h *GiftHandler) GetGiftsReceivedPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	userID, err := uuid.FromString(r.PathValue("userId"))
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	ctx, span := tracer.MyTracer.Start(ctx, "gift_handler.get_gifts_received")
	gifts, total, serviceErr := h.giftService.GetReceived(ctx, userID, page, limit)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	meta := &response.Meta{Page: page, PageSize: limit, Total: total}
	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &gifts, meta, nil))
}
