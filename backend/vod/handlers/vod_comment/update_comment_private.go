package vodcomment

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/shared/pkg/tracer"
	"sen1or/letslive/vod/dto"
	"sen1or/letslive/vod/handlers/utils"
	response "sen1or/letslive/vod/response"

	"github.com/gofrs/uuid/v5"
)

func (h *VODCommentHandler) UpdateCommentPrivateHandler(w http.ResponseWriter, r *http.Request) {
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

	defer r.Body.Close()
	var requestBody dto.UpdateVODCommentRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "update_vod_comment_private_handler")
	comment, serviceErr := h.commentService.UpdateComment(ctx, requestBody, commentId, *userUUID)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, response.FromError(serviceErr))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, comment, nil, nil))
}
