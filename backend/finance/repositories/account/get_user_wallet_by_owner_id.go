package account

import (
	"context"
	"errors"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r postgresAccountRepo) GetUserWalletByOwnerId(ctx context.Context, ownerId uuid.UUID) (*domains.Account, error) {
	query := `
        select id, type, owner_id, status, created_at
        from accounts
        where owner_id = $1 and type = 'user_wallet'
    `
	rows, err := r.dbConn.Query(ctx, query, ownerId)
	if err != nil {
		logger.Errorf(ctx, "db query error [getuserwalletbyownerid: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	account, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Account])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrAccountNotFound
		}
		logger.Errorf(ctx, "db scan error [getuserwalletbyownerid: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &account, nil
}
