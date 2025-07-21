package middlewares

import (
	"net/http"
	"sen1or/letslive/livestream/pkg/logger"
	serviceresponse "sen1or/letslive/livestream/responses"
	"sen1or/letslive/livestream/types"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

func IsAuthorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.URL.Query().Get("userId")
		if len(userId) == 0 {
			logger.Debugf("isAuthor middleware missing userId param")
			writeErrorResponse(w, serviceresponse.ErrInvalidPath)
		}

		accessTokenCookie, err := r.Cookie("ACCESS_TOKEN")
		if err != nil || len(accessTokenCookie.Value) == 0 {
			writeErrorResponse(w, serviceresponse.ErrUnauthorized)
			return
		}

		myClaims := types.MyClaims{}

		// the signature should already been checked from the api gateway before going to this
		_, _, err = jwt.NewParser().ParseUnverified(accessTokenCookie.Value, &myClaims)
		if err != nil {
			writeErrorResponse(w, serviceresponse.ErrUnauthorized)
			return
		}

		_, err = uuid.FromString(myClaims.UserId)
		if err != nil {
			writeErrorResponse(w, serviceresponse.ErrUnauthorized)
			return
		}

		if myClaims.UserId == userId {
			writeErrorResponse(w, serviceresponse.ErrForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
