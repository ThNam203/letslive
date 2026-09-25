package vod

import (
	"context"
	"encoding/json"
	"net/http"
	response "sen1or/letslive/vod/response"

	"github.com/gofrs/uuid/v5"
)

type setAuthorDisabledRequest struct {
	Disabled bool `json:"disabled"`
}

type replaceDisabledAuthorsRequest struct {
	UserIds []uuid.UUID `json:"userIds"`
}

func (h *VODHandler) SetAuthorDisabledInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userId, err := uuid.FromString(r.PathValue("userId"))
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	defer r.Body.Close()
	var reqBody setAuthorDisabledRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	if err := h.vodService.SetAuthorDisabled(ctx, userId, reqBody.Disabled); err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}

func (h *VODHandler) ReplaceDisabledAuthorsInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	defer r.Body.Close()
	var reqBody replaceDisabledAuthorsRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || reqBody.UserIds == nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	if err := h.vodService.ReplaceDisabledAuthors(ctx, reqBody.UserIds); err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}
