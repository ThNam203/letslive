package utils

import (
	"net/http"
	"strconv"
)

const (
	requestPageQueryDefaultMinValue  = 0
	requestLimitQueryDefaultMinValue = 1
	requestLimitQueryDefaultMaxValue = 20
)

func GetPageAndLimitQuery(r *http.Request) (finalPage int, finalLimit int) {
	page := r.URL.Query().Get("page")
	finalPage, pageErr := strconv.Atoi(page)
	if pageErr != nil || finalPage < requestPageQueryDefaultMinValue {
		finalPage = requestPageQueryDefaultMinValue
	}

	limit := r.URL.Query().Get("limit")
	finalLimit, limitErr := strconv.Atoi(limit)
	if limitErr != nil {
		finalLimit = requestLimitQueryDefaultMaxValue
	} else if finalLimit < requestLimitQueryDefaultMinValue {
		finalLimit = requestLimitQueryDefaultMinValue
	} else if finalLimit > requestLimitQueryDefaultMaxValue {
		finalLimit = requestLimitQueryDefaultMaxValue
	}

	return
}
