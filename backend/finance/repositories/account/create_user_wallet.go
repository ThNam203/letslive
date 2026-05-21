package account

import (
	"context"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r postgresAccountRepo) CreateUserWallet(ctx context.Context, ownerId uuid.UUID) (*domains.Account, *response.Response[any]) {
	query := `
        insert into accounts (type, owner_id, status)
        values ('user_wallet', $1, 'active')
        returning id, type, owner_id, status, created_at
    `
	rows, err := r.dbConn.Query(ctx, query, ownerId)
	if err != nil {
		logger.Errorf(ctx, "db query error [createuserwallet: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	account, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Account])
	if err != nil {
		logger.Errorf(ctx, "db scan error [createuserwallet: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &account, nil
}
