package livestream_information

import (
	"context"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresLivestreamInformationRepo) Create(ctx context.Context, userId uuid.UUID) *response.Response[any] {
	result, err := r.dbConn.Exec(ctx, "insert into livestream_information (user_id) values ($1)", userId)
	if err != nil || result.RowsAffected() == 0 {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	return nil
}
