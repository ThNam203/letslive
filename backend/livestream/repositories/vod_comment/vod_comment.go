package vodcomment

import (
	"sen1or/letslive/livestream/domains"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresVODCommentRepo struct {
	db domains.DBTX
}

func NewVODCommentRepository(conn *pgxpool.Pool) domains.VODCommentRepository {
	return &postgresVODCommentRepo{
		db: conn,
	}
}

func (r *postgresVODCommentRepo) WithTx(tx pgx.Tx) domains.VODCommentRepository {
	return &postgresVODCommentRepo{
		db: tx,
	}
}
