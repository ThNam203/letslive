package livestream_information

import (
	"sen1or/letslive/user/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresLivestreamInformationRepo struct {
	dbConn *pgxpool.Pool
}

func NewLivestreamInformationRepository(conn *pgxpool.Pool) domains.LivestreamInformationRepository {
	return &postgresLivestreamInformationRepo{
		dbConn: conn,
	}
}
