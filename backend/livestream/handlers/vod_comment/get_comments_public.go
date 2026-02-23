package vodcomment

import (
	"context"
	"net/http"
	"sen1or/letslive/livestream/handlers/utils"
	"sen1or/letslive/livestream/pkg/tracer"
	response "sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (h *VODCommentHandler) GetCommentsPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	rawVODId := r.PathValue("vodId")
	vodId, err := uuid.FromString(rawVODId)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	page, limit := utils.GetPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_vod_comments_public_handler")
	comments, total, serviceErr := h.commentService.GetCommentsByVODId(ctx, vodId, page, limit)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	meta := &response.Meta{
		Page:     page,
		PageSize: limit,
		Total:    total,
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &comments, meta, nil))
}
