package sign_up_otp

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r *postgresSignUpOTPRepo) GetOTP(ctx context.Context, code string, email string) (*domains.SignUpOTP, error) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT id, code, email, expires_at, created_at, used_at
		FROM sign_up_otps
		WHERE code = $1 AND email = $2
	`, code, email)
	if err != nil {
		logger.Errorf(ctx, "failed to get otp: %s", err)
		return nil, domains.ErrDatabaseQuery
	}
	defer rows.Close()

	otp, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.SignUpOTP])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrSignUpOTPNotFound
		}

		logger.Errorf(ctx, "failed to collect otp: %s", err)
		return nil, domains.ErrDatabaseIssue
	}

	return &otp, nil
}
