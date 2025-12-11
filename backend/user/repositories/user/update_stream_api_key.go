package user

import (
	"context"
	"errors"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresUserRepo) UpdateStreamAPIKey(ctx context.Context, userId uuid.UUID, newKey string) *response.Response[any] {
	result, err := r.dbConn.Exec(ctx, "UPDATE users SET stream_api_key = $1 WHERE id = $2", newKey, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.NewResponseFromTemplate[any](
				response.RES_ERR_USER_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}

		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	} else if result.RowsAffected() == 0 {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_USER_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}

	return nil
}
