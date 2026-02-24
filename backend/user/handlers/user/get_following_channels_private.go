package user

import (
	"context"
	"net/http"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
)

func (h *UserHandler) GetFollowingChannelsPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	authenticatedUserId, cookieErr := utils.GetUserIdFromCookie(r)
	if cookieErr != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_following_channels_private_handler.user_service.get_following_users")
	users, serviceErr := h.userService.GetFollowingUsers(ctx, *authenticatedUserId)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &users, nil, nil))
}
