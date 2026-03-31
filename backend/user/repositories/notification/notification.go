package notification

import (
	"sen1or/letslive/user/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresNotificationRepo struct {
	dbConn *pgxpool.Pool
}

func NewNotificationRepository(conn *pgxpool.Pool) domains.NotificationRepository {
	return &postgresNotificationRepo{
		dbConn: conn,
	}
}
