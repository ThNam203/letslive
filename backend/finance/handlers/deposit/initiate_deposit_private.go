package deposit

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/finance/dto"
	"sen1or/letslive/finance/handlers/utils"
	response "sen1or/letslive/finance/response"
	financeutils "sen1or/letslive/finance/utils"
	"sen1or/letslive/shared/pkg/tracer"
)

func (h *DepositHandler) InitiateDepositPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userId, errResp := utils.GetUserIdFromCookie(r)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	var req dto.DepositRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}
	if err := financeutils.Validator.Struct(req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseWithValidationErrors[any](nil, nil, err))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "initiate_deposit_private_handler.deposit_service.initiate")
	resp, serviceErr := h.depositService.Initiate(ctx, *userId, req)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, resp, nil, nil))
}
