package vodcommentedit

import (
	"sen1or/letslive/vod/domains"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresVODCommentEditRepo struct {
	db domains.DBTX
}

func NewVODCommentEditRepository(conn *pgxpool.Pool) domains.VODCommentEditRepository {
	return &postgresVODCommentEditRepo{
		db: conn,
	}
}

func (r *postgresVODCommentEditRepo) WithTx(tx pgx.Tx) domains.VODCommentEditRepository {
	return &postgresVODCommentEditRepo{
		db: tx,
	}
}
