package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"sen1or/letslive/admin/dto"
	"sen1or/letslive/admin/response"
	"sen1or/letslive/admin/utils"
)

func (h *AuthHandler) LoginPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var req dto.LoginRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}
	if err := utils.Validator.Struct(req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseWithValidationErrors[any](nil, nil, err))
		return
	}

	account, loginErr := h.authService.Login(ctx, req.Email, req.Password)
	if loginErr != nil {
		h.WriteResponse(w, ctx, loginErr)
		return
	}

	accessToken, expiresAt, tokenErr := h.authService.GenerateAccessToken(account.Id.String())
	if tokenErr != nil {
		h.WriteResponse(w, ctx, tokenErr)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "ADMIN_ACCESS_TOKEN",
		Value:    accessToken,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}
