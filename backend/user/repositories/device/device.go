package devicerepo

import (
	"sen1or/letslive/user/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresDeviceRepo struct {
	dbConn *pgxpool.Pool
}

func NewDeviceTokenRepository(conn *pgxpool.Pool) domains.DeviceTokenRepository {
	return &postgresDeviceRepo{dbConn: conn}
}
