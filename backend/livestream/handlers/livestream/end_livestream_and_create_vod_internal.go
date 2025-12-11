package livestream

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/livestream/dto"
	"sen1or/letslive/livestream/pkg/tracer"
	response "sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (h *LivestreamHandler) EndLivestreamAndCreateVODInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	rawStreamId := r.PathValue("livestreamId")
	streamId, err := uuid.FromString(rawStreamId)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}
	defer r.Body.Close()

	var requestBody dto.EndLivestreamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "end_livestream_and_create_vod_internal_handler.livestream_service.end_livestream_and_create_vod")
	serviceErr := h.livestreamService.EndLivestreamAndCreateVOD(ctx, streamId, requestBody)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}
