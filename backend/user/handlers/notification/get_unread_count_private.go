package notification

import (
	"context"
	"net/http"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
)

type unreadCountResponse struct {
	Count int `json:"count"`
}

func (h *NotificationHandler) GetUnreadCountPrivateHandler(w http.ResponseWriter, r *http.Request) {
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

	ctx, span := tracer.MyTracer.Start(ctx, "notification_handler.get_unread_count")
	count, serviceErr := h.notificationService.GetUnreadCount(ctx, userId.String())
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	data := unreadCountResponse{Count: count}
	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(
		response.RES_SUCC_OK,
		&data,
		nil,
		nil,
	))
}
