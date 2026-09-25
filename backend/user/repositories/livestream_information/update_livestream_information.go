package livestream_information

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"

	"github.com/jackc/pgx/v5"
)

func (r *postgresLivestreamInformationRepo) Update(ctx context.Context, livestreamInformation domains.LivestreamInformation) (*domains.LivestreamInformation, error) {
	params := pgx.NamedArgs{
		"user_id":       livestreamInformation.UserID,
		"title":         livestreamInformation.Title,
		"description":   livestreamInformation.Description,
		"thumbnail_url": livestreamInformation.ThumbnailURL,
	}

	rows, err := r.dbConn.Query(ctx, "UPDATE livestream_information SET title = @title, description = @description, thumbnail_url = @thumbnail_url WHERE user_id = @user_id RETURNING *", params)
	if err != nil {
		return nil, domains.ErrDatabaseQuery
	}
	defer rows.Close()

	updatedInformation, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.LivestreamInformation])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrUserNotFound
		}

		return nil, domains.ErrDatabaseIssue
	}

	return &updatedInformation, nil
}
