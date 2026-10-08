package utils

import (
	"net/http"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/finance/types"
	"sen1or/letslive/shared/pkg/jwtauth"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
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

	err = jwtauth.Verify(r.Context(), accessTokenCookie.Value, &myClaims)
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
