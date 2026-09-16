package jwt_token

import (
	"context"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/shared/pkg/logger"
	"time"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresRefreshTokenRepo) RevokeAllTokensOfUser(ctx context.Context, userId uuid.UUID) error {
	var timeNow = time.Now()
	_, err := r.dbConn.Exec(ctx, `
		UPDATE refresh_tokens 
		SET revoked_at = $1 
		WHERE user_id = $2
	`, &timeNow, userId.String())
	if err != nil {
		logger.Errorf(ctx, "failed to revoke refresh tokens of user %s: %s", userId, err)
		return domains.ErrDatabaseQuery
	}

	return nil
}
