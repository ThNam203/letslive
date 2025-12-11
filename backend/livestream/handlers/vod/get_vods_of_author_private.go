package vod

import (
	"context"
	"net/http"
	"sen1or/letslive/livestream/handlers/utils"
	"sen1or/letslive/livestream/pkg/tracer"
	response "sen1or/letslive/livestream/response"
)

func (h *VODHandler) GetVODsOfAuthorPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userUUID, err := utils.GetUserIdFromCookie(r)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	page, limit := utils.GetPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_vods_of_author_private_handler.vod_service.get_all_vods_by_user")
	livestreams, serviceErr := h.vodService.GetAllVODsByUser(ctx, *userUUID, page, limit)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &livestreams, nil, nil))
}
