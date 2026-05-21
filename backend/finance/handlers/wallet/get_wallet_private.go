package wallet

import (
	"context"
	"net/http"
	"sen1or/letslive/finance/handlers/utils"
	response "sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/tracer"
)

func (h *WalletHandler) GetWalletPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userId, errResp := utils.GetUserIdFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_wallet_private_handler.wallet_service.get_or_create")
	wallet, serviceErr := h.walletService.GetOrCreateWallet(ctx, *userId)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, wallet, nil, nil))
}
