package livestream

import (
	"sen1or/letslive/livestream/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresLivestreamRepo struct {
	dbConn *pgxpool.Pool
}

func NewLivestreamRepository(conn *pgxpool.Pool) domains.LivestreamRepository {
	return &postgresLivestreamRepo{
		dbConn: conn,
	}
}
