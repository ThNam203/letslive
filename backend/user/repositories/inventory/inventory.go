package inventoryrepo

import (
	"sen1or/letslive/user/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresInventoryRepo struct {
	dbConn *pgxpool.Pool
}

func NewInventoryRepository(conn *pgxpool.Pool) domains.InventoryRepository {
	return &postgresInventoryRepo{dbConn: conn}
}
