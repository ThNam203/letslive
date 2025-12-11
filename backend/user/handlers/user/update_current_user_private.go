package user

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func (h *UserHandler) UpdateCurrentUserPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	userUUID, cookieErr := utils.GetUserIdFromCookie(r)
	if cookieErr != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}
	defer r.Body.Close()

	var requestBody dto.UpdateUserRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		logger.Errorf(ctx, "failed to decode request body: %s", err)
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "update_current_user_private_handler.user_service.update_user")
	requestBody.Id = uuid.FromStringOrNil(userUUID.String())
	updatedUser, err := h.userService.UpdateUser(ctx, requestBody)
	span.End()

	if err != nil {
		h.WriteResponse(w, ctx, err)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, updatedUser, nil, nil))
}
