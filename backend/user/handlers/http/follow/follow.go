package follow

import (
	"context"
	"net/http"
	"sen1or/letslive/user/handlers/http/basehandler"
	"sen1or/letslive/user/handlers/http/utils"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
	"sen1or/letslive/user/services"
)

type FollowHandler struct {
	basehandler.BaseHandler
	followService services.FollowService
}

func NewFollowHandler(followService services.FollowService) *FollowHandler {
	return &FollowHandler{
		followService: followService,
	}
}

func (h *FollowHandler) FollowPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	followedId := r.PathValue("userId")
	followerId, cookieErr := utils.GetUserIdFromCookie(r)
	if cookieErr != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "follow_private_handler.follow_service.follow")
	serviceErr := h.followService.Follow(ctx, followerId.String(), followedId)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
		response.RES_SUCC_OK,
		nil,
		nil,
		nil,
	))
}
