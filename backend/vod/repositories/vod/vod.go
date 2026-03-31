package vod

import (
	"sen1or/letslive/vod/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresVODRepo struct {
	dbConn *pgxpool.Pool
}

func NewVODRepository(conn *pgxpool.Pool) domains.VODRepository {
	return &postgresVODRepo{
		dbConn: conn,
	}
}
