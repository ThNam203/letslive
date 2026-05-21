package payment

import (
	"context"
	"net/http"
	"sen1or/letslive/finance/handlers/utils"
	response "sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/tracer"

	"github.com/gofrs/uuid/v5"
)

func (h *PaymentHandler) GetPaymentByIdPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userId, errResp := utils.GetUserIdFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	rawId := r.PathValue("paymentId")
	pid, err := uuid.FromString(rawId)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_payment_by_id_private_handler.payment_service.get_for_actor")
	p, serviceErr := h.paymentService.GetForActor(ctx, pid, *userId)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, p, nil, nil))
}
