package user

import (
	"context"
	"net/http"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func (h *UserHandler) GetUserByIdPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	authenticatedUserId, _ := utils.GetUserIdFromCookie(r)
	userId := r.PathValue("userId")
	if len(userId) == 0 {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		))
		return
	}

	userUUID, err := uuid.FromString(userId)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_user_by_id_public_handler.user_service.get_user_public_info_by_id")
	user, serviceErr := h.userService.GetUserPublicInfoById(ctx, userUUID, authenticatedUserId)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, user, nil, nil))
}
