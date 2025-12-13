package jwt_token

import (
	"context"
	serviceresponse "sen1or/letslive/auth/response"
	"time"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresRefreshTokenRepo) RevokeAllTokensOfUser(ctx context.Context, userId uuid.UUID) *serviceresponse.Response[any] {
	var timeNow = time.Now()
	_, err := r.dbConn.Exec(ctx, `
		UPDATE refresh_tokens 
		SET revoked_at = $1 
		WHERE user_id = $2
	`, &timeNow, userId.String())
	if err != nil {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"userId": userId}},
		)
	}

	return nil
}
