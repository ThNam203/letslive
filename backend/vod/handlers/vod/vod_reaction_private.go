package vod

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

func (h *VODHandler) GetMyReactionPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	vodId, userId, ok := h.parseReactionRequest(ctx, w, r)
	if !ok {
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_my_vod_reaction_private_handler")
	result, serviceErr := h.reactionService.GetMyReaction(ctx, vodId, userId)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, response.FromError(serviceErr))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, result, nil, nil))
}

func (h *VODHandler) SetReactionPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	vodId, userId, ok := h.parseReactionRequest(ctx, w, r)
	if !ok {
		return
	}

	defer r.Body.Close()
	var requestBody dto.SetVODReactionRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "set_vod_reaction_private_handler")
	result, serviceErr := h.reactionService.SetReaction(ctx, requestBody, vodId, userId)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, response.FromError(serviceErr))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, result, nil, nil))
}

func (h *VODHandler) RemoveReactionPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	vodId, userId, ok := h.parseReactionRequest(ctx, w, r)
	if !ok {
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "remove_vod_reaction_private_handler")
	result, serviceErr := h.reactionService.RemoveReaction(ctx, vodId, userId)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, response.FromError(serviceErr))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, result, nil, nil))
}

func (h *VODHandler) parseReactionRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	userUUID, cErr := utils.GetUserIdFromCookie(r)
	if cErr != nil {
		h.WriteResponse(w, ctx, cErr)
		return uuid.Nil, uuid.Nil, false
	}

	vodId, err := uuid.FromString(r.PathValue("vodId"))
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return uuid.Nil, uuid.Nil, false
	}

	return vodId, *userUUID, true
}
