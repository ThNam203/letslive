package vod

import (
	"context"
	"net/http"
	"sen1or/letslive/shared/pkg/tracer"
	"sen1or/letslive/vod/handlers/utils"
	response "sen1or/letslive/vod/response"
	"strings"
	"unicode/utf8"
)

const searchQueryMaxLength = 100

func (h *VODHandler) SearchVODsPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if !isValidSearchQuery(query) {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	page, limit := utils.GetPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "search_vods_public_handler.vod_service.search_public_vods")
	vods, serviceErr := h.vodService.SearchPublicVODs(ctx, query, page, limit)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, response.FromError(serviceErr))
		return
	}

	meta := &response.Meta{
		Page:     page,
		PageSize: limit,
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &vods, meta, nil))
}

// Postgres rejects text that isn't valid UTF-8 or contains NUL, which would
// surface as a 500 instead of an input error.
func isValidSearchQuery(query string) bool {
	length := utf8.RuneCountInString(query)
	return length > 0 && length <= searchQueryMaxLength &&
		utf8.ValidString(query) && !strings.ContainsRune(query, 0)
}
