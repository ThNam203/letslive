package repositories

import (
	"context"
	"errors"
	"sen1or/lets-live/user/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LivestreamInformationRepository interface {
	GetByUserId(uuid.UUID) (*domains.LivestreamInformation, error)
	Create(uuid.UUID) error
	Update(domains.LivestreamInformation) (*domains.LivestreamInformation, error)
}

type postgresLivestreamInformationRepo struct {
	dbConn *pgxpool.Pool
}

func NewLivestreamInformationRepository(conn *pgxpool.Pool) LivestreamInformationRepository {
	return &postgresLivestreamInformationRepo{
		dbConn: conn,
	}
}

func (r *postgresLivestreamInformationRepo) GetByUserId(userId uuid.UUID) (*domains.LivestreamInformation, error) {
	rows, err := r.dbConn.Query(context.Background(), "select * from livestream_information where user_id = $1", userId.String())
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.LivestreamInformation])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *postgresLivestreamInformationRepo) Create(userId uuid.UUID) error {
	_, err := r.dbConn.Exec(context.Background(), "insert into livestream_information (user_id) values ($1)", userId)
	if err != nil {
		return err
	}

	return nil
}

func (r *postgresLivestreamInformationRepo) Update(livestreamInformation domains.LivestreamInformation) (*domains.LivestreamInformation, error) {
	params := pgx.NamedArgs{
		"user_id":       livestreamInformation.UserID,
		"title":         livestreamInformation.Title,
		"description":   livestreamInformation.Description,
		"thumbnail_url": livestreamInformation.ThumbnailURL,
	}

	rows, err := r.dbConn.Query(context.Background(), "UPDATE livestream_information SET title = @title, description = @description, thumbnail_url = @thumbnail_url WHERE user_id = @user_id RETURNING *", params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updatedInformation, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.LivestreamInformation])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &updatedInformation, nil
}
