package middlewares

import (
	"net/http"
	"sen1or/letslive/livestream/handlers/utils"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"
	"sen1or/letslive/livestream/types"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

func IsAuthorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userId := r.URL.Query().Get("userId")
		if len(userId) == 0 {
			logger.Debugf(r.Context(), "isAuthor middleware missing userId param")
			utils.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
			return
		}

		accessTokenCookie, err := r.Cookie("ACCESS_TOKEN")
		if err != nil || len(accessTokenCookie.Value) == 0 {
			utils.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
				response.RES_ERR_UNAUTHORIZED,
				nil,
				nil,
				nil,
			))
			return
		}

		myClaims := types.MyClaims{}

		// the signature should already been checked from the api gateway before going to this
		_, _, err = jwt.NewParser().ParseUnverified(accessTokenCookie.Value, &myClaims)
		if err != nil {
			utils.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
				response.RES_ERR_UNAUTHORIZED,
				nil,
				nil,
				nil,
			))
			return
		}

		_, err = uuid.FromString(myClaims.UserId)
		if err != nil {
			utils.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
				response.RES_ERR_UNAUTHORIZED,
				nil,
				nil,
				nil,
			))
			return
		}

		if myClaims.UserId == userId {
			utils.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
				response.RES_ERR_FORBIDDEN,
				nil,
				nil,
				nil,
			))
			return
		}

		next.ServeHTTP(w, r)
	})
}
