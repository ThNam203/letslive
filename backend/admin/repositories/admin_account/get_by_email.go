package adminaccount

import (
	"context"
	"errors"

	"sen1or/letslive/admin/domains"
	"sen1or/letslive/admin/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r *postgresAdminAccountRepo) GetByEmail(ctx context.Context, email string) (*domains.AdminAccount, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT *
		FROM admin_accounts
		WHERE email = $1
	`, email)
	if err != nil {
		logger.Errorf(ctx, "failed to get admin account by email: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY, nil, nil,
			&response.ErrorDetails{response.ErrorDetail{"email": email}},
		)
	}
	defer rows.Close()

	account, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.AdminAccount])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_ADMIN_NOT_FOUND, nil, nil,
				&response.ErrorDetails{response.ErrorDetail{"email": email}},
			)
		}
		logger.Errorf(ctx, "failed to collect row: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE, nil, nil,
			&response.ErrorDetails{response.ErrorDetail{"email": email}},
		)
	}

	return &account, nil
}
