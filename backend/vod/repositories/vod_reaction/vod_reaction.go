package vodreaction

import (
	"sen1or/letslive/vod/domains"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresVODReactionRepo struct {
	db domains.DBTX
}

func NewVODReactionRepository(conn *pgxpool.Pool) domains.VODReactionRepository {
	return &postgresVODReactionRepo{
		db: conn,
	}
}

func (r *postgresVODReactionRepo) WithTx(tx pgx.Tx) domains.VODReactionRepository {
	return &postgresVODReactionRepo{
		db: tx,
	}
}
