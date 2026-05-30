package deposit

import (
	"context"
	"io"
	"net/http"
	"sen1or/letslive/finance/domains"
	response "sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/tracer"
)

// HandleStripeWebhookPublicHandler accepts the Stripe-signed POST. The handler
// has no JWT requirement; signature verification happens inside the service
// via the configured gateway.
func (h *DepositHandler) HandleStripeWebhookPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}
	signature := r.Header.Get("Stripe-Signature")

	ctx, span := tracer.MyTracer.Start(ctx, "handle_stripe_webhook_public_handler.deposit_service.handle_webhook")
	serviceErr := h.depositService.HandleWebhook(ctx, domains.PaymentProviderStripe, payload, signature)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}
