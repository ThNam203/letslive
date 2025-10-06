package repositories

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresLivestreamInformationRepo struct {
	dbConn *pgxpool.Pool
}

func NewLivestreamInformationRepository(conn *pgxpool.Pool) domains.LivestreamInformationRepository {
	return &postgresLivestreamInformationRepo{
		dbConn: conn,
	}
}

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
