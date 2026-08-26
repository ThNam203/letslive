package repositories

import (
	adminaccount "sen1or/letslive/admin/repositories/admin_account"
	"sen1or/letslive/admin/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewAdminAccountRepository(conn *pgxpool.Pool) domains.AdminAccountRepository {
	return adminaccount.NewAdminAccountRepository(conn)
}
