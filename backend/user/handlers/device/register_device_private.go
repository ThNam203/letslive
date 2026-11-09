package devicehandler

import (
	"context"
	"encoding/json"
	"net/http"

	"sen1or/letslive/shared/pkg/tracer"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/response"

	"github.com/go-playground/validator/v10"
)

func (h *DeviceHandler) RegisterDevicePrivateHandler(w http.ResponseWriter, r *http.Request) {
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

	var req dto.RegisterDeviceRequestDTO
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

	ctx, span := tracer.MyTracer.Start(ctx, "device_handler.register_device")
	saved, serviceErr := h.deviceService.Register(ctx, *userId, req)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, response.FromError(serviceErr))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(
		response.RES_SUCC_OK,
		saved,
		nil,
		nil,
	))
}
