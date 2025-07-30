package handlers

import (
	"context"
	"net/http"
	servererrors "sen1or/letslive/user/errors"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/services"
)

type FollowHandler struct {
	ErrorHandler
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
		h.WriteErrorResponse(w, servererrors.ErrUnauthorized)
	}

	ctx, span := tracer.MyTracer.Start(ctx, "follow_private_handler.follow_service.follow")
	serviceErr := h.followService.Follow(ctx, followerId.String(), followedId)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}
	span.End()

	w.WriteHeader(http.StatusCreated)
}

func (h *FollowHandler) UnfollowPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	followedId := r.PathValue("userId")
	followerId, cookieErr := getUserIdFromCookie(r)
	if cookieErr != nil {
		h.WriteErrorResponse(w, servererrors.ErrUnauthorized)
	}

	ctx, span := tracer.MyTracer.Start(ctx, "unfollow_private_handler.follow_service.unfollow")
	serviceErr := h.followService.Unfollow(ctx, followerId.String(), followedId)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}
	span.End()

	w.WriteHeader(http.StatusOK)
}
