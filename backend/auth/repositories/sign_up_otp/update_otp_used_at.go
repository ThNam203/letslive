package sign_up_otp

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/shared/pkg/logger"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresSignUpOTPRepo) UpdateUsedAt(ctx context.Context, otpId uuid.UUID, verifiedAt time.Time) error {
	result, err := r.dbConn.Exec(ctx, `
		UPDATE sign_up_otps
		SET used_at = $1
		WHERE id = $2
	`, verifiedAt, otpId)

	// TODO: test if pgx.ErrNoRows is returned on Exec

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domains.ErrSignUpOTPNotFound
		}

		logger.Errorf(ctx, "failed to update otp used at", err)
		return domains.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return domains.ErrSignUpOTPNotFound
	}

	return nil
}
