package handlers

import (
	"net/http"
	"sen1or/letslive/livestream/constants"
	"sen1or/letslive/livestream/pkg/logger"
	serviceresponse "sen1or/letslive/livestream/responses"
	"sen1or/letslive/livestream/types"
	"strconv"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

func getUserIdFromCookie(r *http.Request) (*uuid.UUID, *serviceresponse.ServiceErrorResponse) {
	accessTokenCookie, err := r.Cookie("ACCESS_TOKEN")
	if err != nil || len(accessTokenCookie.Value) == 0 {
		logger.Debugf("missing credentials")
		return nil, serviceresponse.ErrUnauthorized
	}

	myClaims := types.MyClaims{}

	// the signature should already been checked from the api gateway before going to this
	_, _, err = jwt.NewParser().ParseUnverified(accessTokenCookie.Value, &myClaims)
	if err != nil {
		logger.Debugf("invalid access token: %s", err)
		return nil, serviceresponse.ErrUnauthorized
	}

	userUUID, err := uuid.FromString(myClaims.UserId)
	if err != nil {
		logger.Debugf("userId not valid")
		return nil, serviceresponse.ErrUnauthorized
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
		//h.WriteErrorResponse(w, serviceresponse.ErrMissingLimitParameter)
		finalLimit = constants.REQUEST_LIMIT_QUERY_DEFAULT_MAX_VALUE
	} else if finalLimit < constants.REQUEST_LIMIT_QUERY_DEFAULT_MIN_VALUE {
		finalLimit = constants.REQUEST_LIMIT_QUERY_DEFAULT_MIN_VALUE
	} else if finalLimit > constants.REQUEST_LIMIT_QUERY_DEFAULT_MAX_VALUE {
		finalLimit = constants.REQUEST_LIMIT_QUERY_DEFAULT_MAX_VALUE
	}

	return
}
