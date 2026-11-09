package devicehandler

import (
	"context"
	"net/http"

	"sen1or/letslive/shared/pkg/tracer"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/response"
)

func (h *DeviceHandler) UnregisterDevicePrivateHandler(w http.ResponseWriter, r *http.Request) {
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

	token := r.PathValue("token")

	ctx, span := tracer.MyTracer.Start(ctx, "device_handler.unregister_device")
	serviceErr := h.deviceService.Unregister(ctx, *userId, token)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, response.FromError(serviceErr))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
		response.RES_SUCC_OK,
		nil, nil, nil,
	))
}
