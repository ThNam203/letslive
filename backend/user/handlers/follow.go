package handlers

import (
	"context"
	"net/http"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
	"sen1or/letslive/user/services"
)

type FollowHandler struct {
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
	followerId, cookieErr := getUserIdFromCookie(r)
	if cookieErr != nil {
		writeResponse(w, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
	}

	ctx, span := tracer.MyTracer.Start(ctx, "follow_private_handler.follow_service.follow")
	serviceErr := h.followService.Follow(ctx, followerId.String(), followedId)
	span.End()

	if serviceErr != nil {
		writeResponse(w, serviceErr)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *FollowHandler) UnfollowPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	followedId := r.PathValue("userId")
	followerId, cookieErr := getUserIdFromCookie(r)
	if cookieErr != nil {
		writeResponse(w, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
	}

	ctx, span := tracer.MyTracer.Start(ctx, "unfollow_private_handler.follow_service.unfollow")
	serviceErr := h.followService.Unfollow(ctx, followerId.String(), followedId)
	span.End()

	if serviceErr != nil {
		writeResponse(w, serviceErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}
