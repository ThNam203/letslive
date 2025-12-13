package vod

import (
	"context"
	"net/http"
	"sen1or/letslive/livestream/handlers/utils"
	"sen1or/letslive/livestream/pkg/tracer"
	response "sen1or/letslive/livestream/response"
)

// TODO: recommendation system
func (h *VODHandler) GetRecommendedVODsPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	page, limit := utils.GetPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_recommended_vods_public_handler.vod_service.get_recommended_vods")
	vods, serviceErr := h.vodService.GetRecommendedVODs(ctx, page, limit)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &vods, nil, nil))
}
