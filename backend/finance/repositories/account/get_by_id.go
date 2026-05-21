package account

import (
	"context"
	"errors"
	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r postgresAccountRepo) GetById(ctx context.Context, id uuid.UUID) (*domains.Account, *response.Response[any]) {
	query := `
        select id, type, owner_id, status, created_at
        from accounts
        where id = $1
    `
	rows, err := r.dbConn.Query(ctx, query, id)
	if err != nil {
		logger.Errorf(ctx, "db query error [getaccountbyid: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	account, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Account])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_ACCOUNT_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}
		logger.Errorf(ctx, "db scan error [getaccountbyid: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &account, nil
}
