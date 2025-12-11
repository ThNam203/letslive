package repositories

import (
	"sen1or/letslive/auth/domains"
	authrepo "sen1or/letslive/auth/repositories/auth"
	jwtrepo "sen1or/letslive/auth/repositories/jwt_token"
	otprepo "sen1or/letslive/auth/repositories/sign_up_otp"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewAuthRepository(conn *pgxpool.Pool) domains.AuthRepository {
	return authrepo.NewAuthRepository(conn)
}

func NewRefreshTokenRepository(conn *pgxpool.Pool) domains.RefreshTokenRepository {
	return jwtrepo.NewRefreshTokenRepository(conn)
}

func NewSignUpOTPRepo(conn *pgxpool.Pool) domains.SignUpOTPRepository {
	return otprepo.NewSignUpOTPRepo(conn)
}
