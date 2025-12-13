package vod

import (
	"context"
	"net/http"
	"sen1or/letslive/livestream/handlers/http/utils"
	"sen1or/letslive/livestream/pkg/tracer"
	response "sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (h *VODHandler) DeleteVODPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	rawVODId := r.PathValue("vodId")
	vodId, err := uuid.FromString(rawVODId)

	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	userUUID, cErr := utils.GetUserIdFromCookie(r)
	if cErr != nil {
		h.WriteResponse(w, ctx, cErr)
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "delete_vod_private_handler.vod_service.delete")
	serviceErr := h.vodService.Delete(ctx, vodId, *userUUID)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
