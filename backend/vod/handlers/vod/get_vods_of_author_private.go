package vod

import (
	"context"
	"net/http"
	"sen1or/letslive/shared/pkg/tracer"
	"sen1or/letslive/vod/handlers/utils"
	response "sen1or/letslive/vod/response"
)

func (h *VODHandler) GetVODsOfAuthorPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userUUID, cErr := utils.GetUserIdFromCookie(r)
	if cErr != nil {
		h.WriteResponse(w, ctx, cErr)
		return
	}

	page, limit := utils.GetPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_vods_of_author_private_handler.vod_service.get_all_vods_by_user")
	livestreams, serviceErr := h.vodService.GetAllVODsByUser(ctx, *userUUID, page, limit)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, response.FromError(serviceErr))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &livestreams, nil, nil))
}
