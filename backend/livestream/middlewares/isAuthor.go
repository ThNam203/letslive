package middlewares

import (
	"net/http"
	servererrors "sen1or/lets-live/livestream/errors"
	"sen1or/lets-live/livestream/types"
	"sen1or/lets-live/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

func IsAuthorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.URL.Query().Get("userId")
		if len(userId) == 0 {
			logger.Debugf("isAuthor middleware missing userId param")
			writeErrorResponse(w, servererrors.ErrInvalidPath)
		}

		accessTokenCookie, err := r.Cookie("ACCESS_TOKEN")
		if err != nil || len(accessTokenCookie.Value) == 0 {
			writeErrorResponse(w, servererrors.ErrUnauthorized)
			return
		}

		myClaims := types.MyClaims{}

		// the signature should already been checked from the api gateway before going to this
		_, _, err = jwt.NewParser().ParseUnverified(accessTokenCookie.Value, &myClaims)
		if err != nil {
			writeErrorResponse(w, servererrors.ErrUnauthorized)
			return
		}

		_, err = uuid.FromString(myClaims.UserId)
		if err != nil {
			writeErrorResponse(w, servererrors.ErrUnauthorized)
			return
		}

		if myClaims.UserId == userId {
			writeErrorResponse(w, servererrors.ErrForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
