package auth

import (
	"context"
	"net/http"

	"sen1or/letslive/admin/handlers/middleware"
	"sen1or/letslive/admin/response"
)

type MeResponseDTO struct {
	Email string `json:"email"`
}

func (h *AuthHandler) GetMePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	adminId, ok := r.Context().Value(middleware.AdminIdContextKey).(string)
	if !ok {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_UNAUTHORIZED, nil, nil, nil))
		return
	}

	account, errResp := h.authService.GetByID(ctx, adminId)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &MeResponseDTO{Email: account.Email}, nil, nil))
}
