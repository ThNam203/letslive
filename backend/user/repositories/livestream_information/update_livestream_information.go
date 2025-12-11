package livestream_information

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"

	"github.com/jackc/pgx/v5"
)

func (r *postgresLivestreamInformationRepo) Update(ctx context.Context, livestreamInformation domains.LivestreamInformation) (*domains.LivestreamInformation, *response.Response[any]) {
	params := pgx.NamedArgs{
		"user_id":       livestreamInformation.UserID,
		"title":         livestreamInformation.Title,
		"description":   livestreamInformation.Description,
		"thumbnail_url": livestreamInformation.ThumbnailURL,
	}

	rows, err := r.dbConn.Query(ctx, "UPDATE livestream_information SET title = @title, description = @description, thumbnail_url = @thumbnail_url WHERE user_id = @user_id RETURNING *", params)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	defer rows.Close()

	updatedInformation, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.LivestreamInformation])
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

	return &updatedInformation, nil
}
