package notification

import (
	"context"
	"net/http"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
)

func (h *NotificationHandler) MarkAllAsReadPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	userId, cookieErr := utils.GetUserIdFromCookie(r)
	if cookieErr != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil, nil, nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "notification_handler.mark_all_as_read")
	serviceErr := h.notificationService.MarkAllAsRead(ctx, userId.String())
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
		response.RES_SUCC_OK,
		nil, nil, nil,
	))
}
