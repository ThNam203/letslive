package sign_up_otp

import (
	"context"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r *postgresSignUpOTPRepo) Insert(ctx context.Context, otp domains.SignUpOTP) error {
	_ = pgx.NamedArgs{
		"code":       otp.Code,
		"expires_at": otp.ExpiresAt,
		"email":      otp.Email,
	}

	result, err := r.dbConn.Exec(ctx, `
		INSERT INTO sign_up_otps(code, expires_at, email) 
		VALUES ($1, $2, $3)
	`, otp.Code, otp.ExpiresAt, otp.Email)
	if err != nil {
		logger.Errorf(ctx, "failed to exec insert otp: %s", err)
		return domains.ErrDatabaseQuery
	} else if result.RowsAffected() == 0 {
		logger.Errorf(ctx, "failed to insert otp: %s", err)
		return domains.ErrDatabaseIssue
	}

	return nil
}
