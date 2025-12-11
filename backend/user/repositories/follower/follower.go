package follower

import (
	"sen1or/letslive/user/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresFollowRepo struct {
	dbConn *pgxpool.Pool
}

func NewFollowRepository(conn *pgxpool.Pool) domains.FollowRepository {
	return &postgresFollowRepo{
		dbConn: conn,
	}
}
