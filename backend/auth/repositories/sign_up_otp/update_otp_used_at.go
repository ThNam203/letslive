package sign_up_otp

import (
	"context"
	"errors"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresSignUpOTPRepo) UpdateUsedAt(ctx context.Context, otpId uuid.UUID, verifiedAt time.Time) *serviceresponse.Response[any] {
	result, err := r.dbConn.Exec(ctx, `
		UPDATE sign_up_otps
		SET used_at = $1
		WHERE id = $2
	`, verifiedAt, otpId)

	// TODO: test if pgx.ErrNoRows is returned on Exec

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return serviceresponse.NewResponseFromTemplate[any](
				serviceresponse.RES_ERR_SIGN_UP_OTP_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}

		logger.Errorf(ctx, "failed to update otp used at", err)
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	if result.RowsAffected() == 0 {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_SIGN_UP_OTP_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}

	return nil
}
