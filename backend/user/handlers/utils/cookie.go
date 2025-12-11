package utils

import (
	"net/http"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/response"
	"sen1or/letslive/user/types"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

func GetUserIdFromCookie(r *http.Request) (*uuid.UUID, *response.Response[any]) {
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
