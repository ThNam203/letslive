package shopitemrepo

import (
	"sen1or/letslive/finance/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresShopItemRepo struct {
	dbConn *pgxpool.Pool
}

func NewShopItemRepository(conn *pgxpool.Pool) domains.ShopItemRepository {
	return &postgresShopItemRepo{dbConn: conn}
}
