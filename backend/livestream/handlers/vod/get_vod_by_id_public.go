package vod

import (
	"context"
	"net/http"
	"sen1or/letslive/livestream/pkg/tracer"
	response "sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (h *VODHandler) GetVODByIdPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	streamId := r.PathValue("vodId")
	if len(streamId) == 0 {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	vodUUID, err := uuid.FromString(streamId)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_vod_by_id_public_handler.vod_service.get_vod_by_id")
	vod, serviceErr := h.vodService.GetVODById(ctx, vodUUID)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, vod, nil, nil))
}
