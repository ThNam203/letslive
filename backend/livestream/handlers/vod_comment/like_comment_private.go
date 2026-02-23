package vodcomment

import (
	"context"
	"net/http"
	"sen1or/letslive/livestream/handlers/utils"
	"sen1or/letslive/livestream/pkg/tracer"
	response "sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (h *VODCommentHandler) LikeCommentPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	rawCommentId := r.PathValue("commentId")
	commentId, err := uuid.FromString(rawCommentId)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	userUUID, cErr := utils.GetUserIdFromCookie(r)
	if cErr != nil {
		h.WriteResponse(w, ctx, cErr)
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "like_vod_comment_private_handler")
	serviceErr := h.commentService.LikeComment(ctx, commentId, *userUUID)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}
