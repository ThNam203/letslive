package user

import (
	"context"
	"net/http"
	"sen1or/letslive/shared/pkg/tracer"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/response"
	"strconv"
	"strings"
	"unicode/utf8"
)

// paging is optional so existing callers keep getting the first 10 results
const (
	searchUsersDefaultLimit   = 10
	searchUsersMaxLimit       = 20
	searchUsersMaxQueryLength = 100
	searchUsersMaxPage        = 10000
)

func (h *UserHandler) SearchUsersPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	authenticatedUserId, _ := utils.GetUserIdFromCookie(r)
	username := r.URL.Query().Get("username")

	if utf8.RuneCountInString(username) > searchUsersMaxQueryLength || !utf8.ValidString(username) || strings.ContainsRune(username, 0) {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		))
		return
	}

	page, limit := getSearchUsersPageAndLimit(r)

	ctx, span := tracer.MyTracer.Start(ctx, "search_users_public_handler.user_service.search_users_by_username")
	users, err := h.userService.SearchUsersByUsername(ctx, username, authenticatedUserId, page, limit)
	span.End()

	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	meta := &response.Meta{
		Page:     page,
		PageSize: limit,
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &users, meta, nil))
}

func getSearchUsersPageAndLimit(r *http.Request) (int, int) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 0 {
		page = 0
	}
	page = min(page, searchUsersMaxPage)

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = searchUsersDefaultLimit
	}
	limit = min(max(limit, 1), searchUsersMaxLimit)

	return page, limit
}
