package handlers

import (
	"encoding/json"
	"net/http"
	"sen1or/letslive/livestream/constants"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"
	"sen1or/letslive/livestream/types"
	"strconv"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

func WriteResponse(w http.ResponseWriter, res *response.Response[any]) {
	res.RequestId = w.Header().Get("requestId")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(res.StatusCode)
	json.NewEncoder(w).Encode(res)
}

func getUserIdFromCookie(r *http.Request) (*uuid.UUID, *response.Response[any]) {
	accessTokenCookie, err := r.Cookie("ACCESS_TOKEN")
	if err != nil || len(accessTokenCookie.Value) == 0 {
		logger.Debugf(r.Context(), "missing credentials")
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		)
	}

	myClaims := types.MyClaims{}

	// the signature should already been checked from the api gateway before going to this
	_, _, err = jwt.NewParser().ParseUnverified(accessTokenCookie.Value, &myClaims)
	if err != nil {
		logger.Debugf(r.Context(), "invalid access token: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		)
	}

	userUUID, err := uuid.FromString(myClaims.UserId)
	if err != nil {
		logger.Debugf(r.Context(), "userId not valid")
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		)
	}

	return &userUUID, nil
}

func getPageAndLimitQuery(r *http.Request) (finalPage int, finalLimit int) {
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
