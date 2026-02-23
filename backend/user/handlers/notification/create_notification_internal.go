package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"

	"github.com/go-playground/validator/v10"
)

func (h *NotificationHandler) CreateNotificationInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var req dto.CreateNotificationRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil, nil, nil,
		))
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseWithValidationErrors[any](nil, nil, err))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "notification_handler.create_notification")
	created, serviceErr := h.notificationService.CreateNotification(ctx, req)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(
		response.RES_SUCC_OK,
		created,
		nil,
		nil,
	))
}
