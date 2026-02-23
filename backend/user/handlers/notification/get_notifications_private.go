package notification

import (
	"context"
	"net/http"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
	"strconv"
)

func (h *NotificationHandler) GetNotificationsPrivateHandler(w http.ResponseWriter, r *http.Request) {
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

	page := 0
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed >= 0 {
			page = parsed
		}
	}

	ctx, span := tracer.MyTracer.Start(ctx, "notification_handler.get_notifications")
	notifications, total, serviceErr := h.notificationService.GetNotifications(ctx, userId.String(), page)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	meta := &response.Meta{
		Page:     page,
		PageSize: 20,
		Total:    total,
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(
		response.RES_SUCC_OK,
		&notifications,
		meta,
		nil,
	))
}
