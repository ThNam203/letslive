package handlers

import (
	"net/http"
	servererrors "sen1or/letslive/user/errors"
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

func (h *FollowHandler) FollowHandler(w http.ResponseWriter, r *http.Request) {
	followedId := r.PathValue("userId")
	followerId, cookieErr := getUserIdFromCookie(r)
	if cookieErr != nil {
		h.WriteErrorResponse(w, servererrors.ErrUnauthorized)
	}

	serviceErr := h.followService.Follow(followerId.String(), followedId)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *FollowHandler) UnfollowHandler(w http.ResponseWriter, r *http.Request) {
	followedId := r.PathValue("userId")
	followerId, cookieErr := getUserIdFromCookie(r)
	if cookieErr != nil {
		h.WriteErrorResponse(w, servererrors.ErrUnauthorized)
	}

	serviceErr := h.followService.Unfollow(followerId.String(), followedId)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}
