package livestream

import (
	"context"
	"net/http"
	"sen1or/letslive/livestream/handlers/http/utils"
	"sen1or/letslive/livestream/pkg/tracer"
	response "sen1or/letslive/livestream/response"
)

func (h *LivestreamHandler) GetRecommendedLivestreamsPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	page, limit := utils.GetPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_recommended_livestreams_public_handler.livestream_service.get_recommended_livestreams")
	livestreams, serviceErr := h.livestreamService.GetRecommendedLivestreams(ctx, page, limit)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(
		response.RES_SUCC_OK,
		&livestreams,
		nil,
		nil,
	))
}
