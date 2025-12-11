package livestream_information

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresLivestreamInformationRepo) GetByUserId(ctx context.Context, userId uuid.UUID) (*domains.LivestreamInformation, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, "select * from livestream_information where user_id = $1", userId.String())
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.LivestreamInformation])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_USER_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}

		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return &user, nil
}
