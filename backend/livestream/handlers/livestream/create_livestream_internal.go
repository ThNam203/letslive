package livestream

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/livestream/dto"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/pkg/tracer"
	response "sen1or/letslive/livestream/response"
)

func (h *LivestreamHandler) CreateLivestreamInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	var body dto.CreateLivestreamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logger.Debugf(ctx, "create livestream body invalid: %s", err.Error())
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "create_livestream_internal_handler.livestream_service.create")
	createdLivestream, err := h.livestreamService.Create(ctx, body)
	span.End()

	if err != nil {
		h.WriteResponse(w, ctx, err)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, createdLivestream, nil, nil))
}
