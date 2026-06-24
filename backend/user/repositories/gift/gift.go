package giftrepo

import (
	"sen1or/letslive/user/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresGiftRepo struct {
	dbConn *pgxpool.Pool
}

func NewGiftRepository(conn *pgxpool.Pool) domains.GiftRepository {
	return &postgresGiftRepo{dbConn: conn}
}
