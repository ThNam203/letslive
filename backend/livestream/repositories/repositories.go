package repositories

import (
	"sen1or/letslive/livestream/domains"
	livestreamrepo "sen1or/letslive/livestream/repositories/livestream"
	vodrepo "sen1or/letslive/livestream/repositories/vod"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewLivestreamRepository(conn *pgxpool.Pool) domains.LivestreamRepository {
	return livestreamrepo.NewLivestreamRepository(conn)
}

func NewVODRepository(conn *pgxpool.Pool) domains.VODRepository {
	return vodrepo.NewVODRepository(conn)
}
