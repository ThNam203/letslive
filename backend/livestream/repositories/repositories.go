package repositories

import (
	"sen1or/letslive/livestream/domains"
	livestreamrepo "sen1or/letslive/livestream/repositories/livestream"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewLivestreamRepository(conn *pgxpool.Pool) domains.LivestreamRepository {
	return livestreamrepo.NewLivestreamRepository(conn)
}
