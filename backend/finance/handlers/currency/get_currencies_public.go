package currency

import (
	"context"
	"net/http"
	response "sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/tracer"
)

func (h *CurrencyHandler) GetCurrenciesPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	ctx, span := tracer.MyTracer.Start(ctx, "get_currencies_public_handler.currency_service.list")
	currencies, serviceErr := h.currencyService.List(ctx)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &currencies, nil, nil))
}
