package user

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
)

func (h *UserHandler) CreateUserInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var body dto.CreateUserRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "create_user_internal_handler.user_service.create_new_user")
	createdUser, err := h.userService.CreateNewUser(ctx, body)
	span.End()

	if err != nil {
		h.WriteResponse(w, ctx, err)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, createdUser, nil, nil))
}
