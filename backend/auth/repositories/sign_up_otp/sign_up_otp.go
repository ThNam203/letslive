package sign_up_otp

import (
	"sen1or/letslive/auth/domains"

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
