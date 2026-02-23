package vodcomment

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

func (h *VODCommentHandler) CreateCommentPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userUUID, cErr := utils.GetUserIdFromCookie(r)
	if cErr != nil {
		h.WriteResponse(w, ctx, cErr)
		return
	}

	rawVODId := r.PathValue("vodId")
	vodId, err := uuid.FromString(rawVODId)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	defer r.Body.Close()
	var requestBody dto.CreateVODCommentRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "create_vod_comment_private_handler")
	comment, serviceErr := h.commentService.CreateComment(ctx, requestBody, vodId, *userUUID)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, comment, nil, nil))
}
