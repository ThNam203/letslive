package vod

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/vod/dto"
	"sen1or/letslive/shared/pkg/tracer"
	response "sen1or/letslive/vod/response"

	"github.com/gofrs/uuid/v5"
)

func (h *VODHandler) RegisterViewPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	vodIdStr := r.PathValue("vodId")
	if len(vodIdStr) == 0 {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	vodUUID, err := uuid.FromString(vodIdStr)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	var body dto.RegisterViewRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	if err := validate.Struct(body); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseWithValidationErrors[any](nil, nil, err))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "register_view_public_handler.vod_service.register_view")
	serviceErr := h.vodService.RegisterView(ctx, vodUUID, body.WatchedSeconds)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}
