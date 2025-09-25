package repositories

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresSignUpOTPRepo struct {
	dbConn *pgxpool.Pool
}

func NewSignUpOTPRepo(conn *pgxpool.Pool) domains.SignUpOTPRepository {
	return &postgresSignUpOTPRepo{
		dbConn: conn,
	}
}

func (r *postgresSignUpOTPRepo) Insert(ctx context.Context, otp domains.SignUpOTP) *serviceresponse.Response[any] {
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
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	} else if result.RowsAffected() == 0 {
		logger.Errorf(ctx, "failed to insert otp: %s", err)
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return nil
}

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
