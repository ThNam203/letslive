package user

import (
	"context"
	"net/http"
	"sen1or/letslive/user/handlers/http/utils"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
)

func (h *UserHandler) SearchUsersPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	authenticatedUserId, _ := utils.GetUserIdFromCookie(r)
	username := r.URL.Query().Get("username")

	ctx, span := tracer.MyTracer.Start(ctx, "search_users_public_handler.user_service.search_users_by_username")
	users, err := h.userService.SearchUsersByUsername(ctx, username, authenticatedUserId)
	span.End()

	if err != nil {
		h.WriteResponse(w, ctx, err)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &users, nil, nil))
}
