package user

import (
	"context"
	"net/http"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
	"strconv"
)

func (h *UserHandler) GetRecommendedChannelsPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 0 {
		page = 0
	}

	authenticatedUserId, _ := utils.GetUserIdFromCookie(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_recommended_channels_public_handler.user_service.get_recommended_users")
	users, serviceErr := h.userService.GetRecommendedUsers(ctx, authenticatedUserId, page)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &users, nil, nil))
}
