package sign_up_otp

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"

	"github.com/jackc/pgx/v5"
)

func (r *postgresSignUpOTPRepo) GetOTP(ctx context.Context, code string, email string) (*domains.SignUpOTP, *serviceresponse.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT id, code, email, expires_at, created_at, used_at
		FROM sign_up_otps
		WHERE code = $1 AND email = $2
	`, code, email)
	if err != nil {
		logger.Errorf(ctx, "failed to get otp: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	defer rows.Close()

	otp, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.SignUpOTP])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serviceresponse.NewResponseFromTemplate[any](
				serviceresponse.RES_ERR_SIGN_UP_OTP_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}

		logger.Errorf(ctx, "failed to collect otp: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			&serviceresponse.ErrorDetails{serviceresponse.ErrorDetail{"code": code, "email": email}},
		)
	}

	return &otp, nil
}
