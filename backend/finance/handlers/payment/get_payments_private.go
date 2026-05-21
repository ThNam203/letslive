package payment

import (
	"context"
	"net/http"
	"sen1or/letslive/finance/handlers/utils"
	response "sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/tracer"
)

func (h *PaymentHandler) GetPaymentsPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userId, errResp := utils.GetUserIdFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	page, limit := utils.GetPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_payments_private_handler.payment_service.list_for_actor")
	payments, total, serviceErr := h.paymentService.ListForActor(ctx, *userId, page, limit)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	meta := &response.Meta{
		Page:     page,
		PageSize: limit,
		Total:    total,
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &payments, meta, nil))
}
