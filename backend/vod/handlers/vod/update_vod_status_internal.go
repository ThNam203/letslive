package vod

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/vod/domains"
	"sen1or/letslive/shared/pkg/tracer"
	response "sen1or/letslive/vod/response"

	"github.com/gofrs/uuid/v5"
)

type UpdateVODStatusRequest struct {
	Status       string  `json:"status"`
	PlaybackUrl  *string `json:"playbackUrl,omitempty"`
	ThumbnailUrl *string `json:"thumbnailUrl,omitempty"`
}

func (h *VODHandler) UpdateVODStatusInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	rawVodId := r.PathValue("vodId")
	vodId, err := uuid.FromString(rawVodId)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	defer r.Body.Close()

	var reqBody UpdateVODStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	vodStatus := domains.VODStatus(reqBody.Status)

	ctx, span := tracer.MyTracer.Start(ctx, "update_vod_status_internal_handler.vod_service.update_vod_status")
	serviceErr := h.vodService.UpdateVODStatus(ctx, vodId, vodStatus, reqBody.PlaybackUrl, reqBody.ThumbnailUrl)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}
