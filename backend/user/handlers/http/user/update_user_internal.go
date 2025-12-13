package user

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func (h *UserHandler) UpdateUserInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	userID := r.PathValue("userId")
	defer r.Body.Close()

	if len(userID) == 0 {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		))
		return
	}

	var requestBody dto.UpdateUserRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}
	requestBody.Id = uuid.FromStringOrNil(userID)

	ctx, span := tracer.MyTracer.Start(ctx, "update_user_internal_handler.user_service.update_user_internal")
	updatedUser, err := h.userService.UpdateUserInternal(ctx, requestBody)
	span.End()

	if err != nil {
		h.WriteResponse(w, ctx, err)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, updatedUser, nil, nil))
}
