package vodcommentlike

import (
	"sen1or/letslive/livestream/domains"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresVODCommentLikeRepo struct {
	db domains.DBTX
}

func NewVODCommentLikeRepository(conn *pgxpool.Pool) domains.VODCommentLikeRepository {
	return &postgresVODCommentLikeRepo{
		db: conn,
	}
}

func (r *postgresVODCommentLikeRepo) WithTx(tx pgx.Tx) domains.VODCommentLikeRepository {
	return &postgresVODCommentLikeRepo{
		db: tx,
	}
}
