package repositories

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"
	servererrors "sen1or/letslive/user/errors"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LivestreamInformationRepository interface {
	GetByUserId(uuid.UUID) (*domains.LivestreamInformation, *servererrors.ServerError)
	Create(uuid.UUID) *servererrors.ServerError
	Update(domains.LivestreamInformation) (*domains.LivestreamInformation, *servererrors.ServerError)
}

type postgresLivestreamInformationRepo struct {
	dbConn *pgxpool.Pool
}

func NewLivestreamInformationRepository(conn *pgxpool.Pool) LivestreamInformationRepository {
	return &postgresLivestreamInformationRepo{
		dbConn: conn,
	}
}

func (r *postgresLivestreamInformationRepo) GetByUserId(userId uuid.UUID) (*domains.LivestreamInformation, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), "select * from livestream_information where user_id = $1", userId.String())
	if err != nil {
		return nil, servererrors.ErrDatabaseQuery
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.LivestreamInformation])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrUserNotFound
		}

		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresLivestreamInformationRepo) Create(userId uuid.UUID) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), "insert into livestream_information (user_id) values ($1)", userId)
	if err != nil || result.RowsAffected() == 0 {
		return servererrors.ErrDatabaseQuery
	}

	return nil
}

func (r *postgresLivestreamInformationRepo) Update(livestreamInformation domains.LivestreamInformation) (*domains.LivestreamInformation, *servererrors.ServerError) {
	params := pgx.NamedArgs{
		"user_id":       livestreamInformation.UserID,
		"title":         livestreamInformation.Title,
		"description":   livestreamInformation.Description,
		"thumbnail_url": livestreamInformation.ThumbnailURL,
	}

	rows, err := r.dbConn.Query(context.Background(), "UPDATE livestream_information SET title = @title, description = @description, thumbnail_url = @thumbnail_url WHERE user_id = @user_id RETURNING *", params)
	if err != nil {
		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	updatedInformation, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.LivestreamInformation])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrUserNotFound
		}

		return nil, servererrors.ErrDatabaseIssue
	}

	return &updatedInformation, nil
}
