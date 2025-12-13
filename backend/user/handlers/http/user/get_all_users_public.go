package user

import (
	"context"
	"net/http"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
	"strconv"
)

func (h *UserHandler) GetAllUsersPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	page, err := strconv.Atoi(r.URL.Query().Get("page"))

	if err != nil || page < 0 {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_all_users_public_handler.user_service.get_all_users")
	users, serviceErr := h.userService.GetAllUsers(ctx, page)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &users, nil, nil))
}
