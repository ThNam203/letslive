package livestream

import (
	"context"
	"net/http"
	"sen1or/letslive/livestream/handlers/utils"
	response "sen1or/letslive/livestream/response"
	"sen1or/letslive/shared/pkg/tracer"
	"strings"
	"unicode/utf8"
)

const searchQueryMaxLength = 100

func (h *LivestreamHandler) SearchLivestreamsPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if !isValidSearchQuery(query) {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	page, limit := utils.GetPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "search_livestreams_public_handler.livestream_service.search_live_livestreams")
	livestreams, serviceErr := h.livestreamService.SearchLiveLivestreams(ctx, query, page, limit)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, response.FromError(serviceErr))
		return
	}

	meta := &response.Meta{
		Page:     page,
		PageSize: limit,
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &livestreams, meta, nil))
}

// Postgres rejects text that isn't valid UTF-8 or contains NUL, which would
// surface as a 500 instead of an input error.
func isValidSearchQuery(query string) bool {
	length := utf8.RuneCountInString(query)
	return length > 0 && length <= searchQueryMaxLength &&
		utf8.ValidString(query) && !strings.ContainsRune(query, 0)
}
