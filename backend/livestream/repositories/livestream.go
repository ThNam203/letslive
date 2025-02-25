package repositories

import (
	"context"
	"errors"
	"sen1or/lets-live/livestream/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LivestreamRepository interface {
	GetById(uuid.UUID) (*domains.Livestream, error)
	GetByUser(uuid.UUID) ([]domains.Livestream, error)

	Create(domains.Livestream) (*domains.Livestream, error)
	Update(domains.Livestream) (*domains.Livestream, error)
	Delete(uuid.UUID) error
}

type postgresLivestreamRepo struct {
	dbConn *pgxpool.Pool
}

func NewLivestreamRepository(conn *pgxpool.Pool) LivestreamRepository {
	return &postgresLivestreamRepo{
		dbConn: conn,
	}
}

func (r *postgresLivestreamRepo) GetById(userId uuid.UUID) (*domains.Livestream, error) {
	rows, err := r.dbConn.Query(context.Background(), "select * from livestreams where id = $1", userId.String())
	if err != nil {
		return nil, err
	}

	livestream, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Livestream])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &livestream, nil
}

func (r *postgresLivestreamRepo) GetByUser(userId uuid.UUID) ([]domains.Livestream, error) {
	rows, err := r.dbConn.Query(context.Background(), `
		SELECT * 
		FROM livestreams 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	livestreams, err := pgx.CollectRows(rows, pgx.RowToStructByName[domains.Livestream])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return livestreams, nil
}

func (r *postgresLivestreamRepo) Create(newLivestream domains.Livestream) (*domains.Livestream, error) {
	params := pgx.NamedArgs{
		"title":         newLivestream.Title,
		"user_id":       newLivestream.UserId,
		"description":   newLivestream.Description,
		"thumbnail_url": newLivestream.ThumbnailURL,
		"status":        newLivestream.Status,
		"view_count":    0,
		//"ended_at": nil,
		"playback_url": newLivestream.PlaybackURL,
	}

	rows, err := r.dbConn.Query(context.Background(), "insert into livestreams (title, user_id, description, thumbnail_url, status, view_count, playback_url) values (@title, @user_id, @description, @thumbnail_url, @status, @view_count, @playback_url) returning *", params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Livestream])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &user, err
}

func (r *postgresLivestreamRepo) Update(livestream domains.Livestream) (*domains.Livestream, error) {
	rows, err := r.dbConn.Query(
		context.Background(),
		`UPDATE livestreams 
		 SET title = $1, description = $2, status = $3, view_count = $4, thumbnail_url = $5, playback_url = $6, ended_at = $7, updated_at = NOW()
		 WHERE id = $8
		 RETURNING *`,
		livestream.Title, livestream.Description, livestream.Status, livestream.ViewCount, livestream.ThumbnailURL, livestream.PlaybackURL, livestream.EndedAt, livestream.Id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updatedLivestream, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Livestream])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &updatedLivestream, err
}

func (r *postgresLivestreamRepo) Delete(livestreamId uuid.UUID) error {
	_, err := r.dbConn.Exec(context.Background(), "DELETE FROM livestreams WHERE id = $1", livestreamId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRecordNotFound
		}

		return err // ?
	}

	return err
}
