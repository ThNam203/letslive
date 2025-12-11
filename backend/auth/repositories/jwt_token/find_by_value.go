package jwt_token

import (
	"context"
	"errors"
	"sen1or/letslive/auth/domains"
	serviceresponse "sen1or/letslive/auth/response"

	"github.com/jackc/pgx/v5"
)

func (r *postgresRefreshTokenRepo) FindByValue(ctx context.Context, tokenVal string) (*domains.RefreshToken, *serviceresponse.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM refresh_tokens 
		WHERE token = $1
	`, tokenVal)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serviceresponse.NewResponseFromTemplate[any](
				serviceresponse.RES_ERR_REFRESH_TOKEN_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	defer rows.Close()

	token, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.RefreshToken])
	if err != nil {
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return &token, nil
}
