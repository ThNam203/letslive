package vod

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/livestream/dto"
	"sen1or/letslive/livestream/handlers/utils"
	"sen1or/letslive/livestream/pkg/tracer"
	response "sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (h *VODHandler) UpdateVODMetadataPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userId, err := utils.GetUserIdFromCookie(r)
	if err != nil {
		h.WriteResponse(w, ctx, err)
		return
	}

	rawStreamId := r.PathValue("vodId")
	streamId, er := uuid.FromString(rawStreamId)
	if er != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}
	defer r.Body.Close()

	var requestBody dto.UpdateVODRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "update_vod_metadata_private_handler.vod_service.update_vod_metadata")
	updatedVOD, serviceErr := h.vodService.UpdateVODMetadata(ctx, requestBody, streamId, *userId)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, updatedVOD, nil, nil))
}
