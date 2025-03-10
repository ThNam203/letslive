package repositories

import (
	"context"
	"errors"
	"sen1or/lets-live/livestream/domains"
	servererrors "sen1or/lets-live/livestream/errors"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LivestreamRepository interface {
	GetById(uuid.UUID) (*domains.Livestream, *servererrors.ServerError)
	GetAllLivestreamings(page int) ([]domains.Livestream, *servererrors.ServerError)
	GetByUser(uuid.UUID) ([]domains.Livestream, *servererrors.ServerError)

	CheckIsUserLivestreaming(uuid.UUID) (bool, *servererrors.ServerError)

	Create(domains.Livestream) (*domains.Livestream, *servererrors.ServerError)
	Update(domains.Livestream) (*domains.Livestream, *servererrors.ServerError)
	Delete(uuid.UUID) *servererrors.ServerError
}

type postgresLivestreamRepo struct {
	dbConn *pgxpool.Pool
}

func NewLivestreamRepository(conn *pgxpool.Pool) LivestreamRepository {
	return &postgresLivestreamRepo{
		dbConn: conn,
	}
}

func (r *postgresLivestreamRepo) GetById(userId uuid.UUID) (*domains.Livestream, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), "select * from livestreams where id = $1", userId.String())
	if err != nil {
		return nil, servererrors.ErrDatabaseQuery
	}

	livestream, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Livestream])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrLivestreamNotFound
		}

		return nil, servererrors.ErrDatabaseIssue
	}

	return &livestream, nil
}

func (r *postgresLivestreamRepo) GetByUser(userId uuid.UUID) ([]domains.Livestream, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), `
		SELECT * 
		FROM livestreams 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userId)
	if err != nil {
		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	livestreams, err := pgx.CollectRows(rows, pgx.RowToStructByName[domains.Livestream])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrLivestreamNotFound
		}
		return nil, servererrors.ErrDatabaseIssue
	}

	return livestreams, nil
}

func (r *postgresLivestreamRepo) GetAllLivestreamings(page int) ([]domains.Livestream, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), "select * from livestreams where ended_at IS NULL OFFSET $1 LIMIT $2", 10*page, 10)
	if err != nil {
		return nil, servererrors.ErrDatabaseQuery
	}

	livestreams, err := pgx.CollectRows(rows, pgx.RowToStructByName[domains.Livestream])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrLivestreamNotFound
		}

		return nil, servererrors.ErrDatabaseIssue
	}

	return livestreams, nil
}

func (r *postgresLivestreamRepo) CheckIsUserLivestreaming(userId uuid.UUID) (bool, *servererrors.ServerError) {
	var exists bool
	err := r.dbConn.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM livestreams WHERE ended_at IS NULL and user_id = $1)", userId).Scan(&exists)
	if err != nil {
		return false, servererrors.ErrDatabaseQuery
	}

	return exists, nil
}

func (r *postgresLivestreamRepo) Create(newLivestream domains.Livestream) (*domains.Livestream, *servererrors.ServerError) {
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
		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Livestream])
	if err != nil {
		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresLivestreamRepo) Update(livestream domains.Livestream) (*domains.Livestream, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(
		context.Background(),
		`UPDATE livestreams 
		 SET title = $1, description = $2, status = $3, view_count = $4, thumbnail_url = $5, playback_url = $6, ended_at = $7, duration = $8, updated_at = NOW()
		 WHERE id = $9
		 RETURNING *`,
		livestream.Title, livestream.Description, livestream.Status, livestream.ViewCount, livestream.ThumbnailURL, livestream.PlaybackURL, livestream.EndedAt, livestream.Duration, livestream.Id,
	)
	if err != nil {
		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	updatedLivestream, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Livestream])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrLivestreamNotFound
		}

		return nil, servererrors.ErrDatabaseIssue
	}

	return &updatedLivestream, nil
}

func (r *postgresLivestreamRepo) Delete(livestreamId uuid.UUID) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), "DELETE FROM livestreams WHERE id = $1", livestreamId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return servererrors.ErrLivestreamNotFound
		}

		return servererrors.ErrDatabaseIssue
	}

	if result.RowsAffected() == 0 {
		return servererrors.ErrLivestreamNotFound
	}

	return nil
}
