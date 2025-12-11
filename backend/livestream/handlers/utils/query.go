package utils

import (
	"net/http"
	"sen1or/letslive/livestream/constants"
	"strconv"
)

func GetPageAndLimitQuery(r *http.Request) (finalPage int, finalLimit int) {
	page := r.URL.Query().Get("page")
	finalPage, pageErr := strconv.Atoi(page)
	if pageErr != nil || finalPage < constants.REQUEST_PAGE_QUERY_DEFAULT_MIN_VALUE {
		finalPage = constants.REQUEST_PAGE_QUERY_DEFAULT_MIN_VALUE
	}

	limit := r.URL.Query().Get("limit")
	finalLimit, limitErr := strconv.Atoi(limit)
	if limitErr != nil {
		finalLimit = constants.REQUEST_LIMIT_QUERY_DEFAULT_MAX_VALUE
	} else if finalLimit < constants.REQUEST_LIMIT_QUERY_DEFAULT_MIN_VALUE {
		finalLimit = constants.REQUEST_LIMIT_QUERY_DEFAULT_MIN_VALUE
	} else if finalLimit > constants.REQUEST_LIMIT_QUERY_DEFAULT_MAX_VALUE {
		finalLimit = constants.REQUEST_LIMIT_QUERY_DEFAULT_MAX_VALUE
	}

	return
}
