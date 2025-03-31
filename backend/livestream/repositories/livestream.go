package repositories

import (
	"context"
	"errors"
	"sen1or/letslive/livestream/domains"
	servererrors "sen1or/letslive/livestream/errors"
	"sen1or/letslive/livestream/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresLivestreamRepo struct {
	dbConn *pgxpool.Pool
}

func NewLivestreamRepository(conn *pgxpool.Pool) domains.LivestreamRepository {
	return &postgresLivestreamRepo{
		dbConn: conn,
	}
}

func (r postgresLivestreamRepo) GetById(userId uuid.UUID) (*domains.Livestream, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), `
		SELECT *
		FROM livestreams 
		WHERE id = $1
	`, userId.String())
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

// we dont care if the result is good or not
func (r postgresLivestreamRepo) AddOneToViewCount(livestreamId uuid.UUID) {
	_, err := r.dbConn.Exec(context.Background(), `
		UPDATE livestreams 
		SET view_count = view_count + 1
		WHERE id = $1
	`, livestreamId)

	if err != nil {
		logger.Errorf("failed to add one to view count: %s", err)
	}
}

func (r postgresLivestreamRepo) GetByUser(userId uuid.UUID) ([]domains.Livestream, *servererrors.ServerError) {
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

func (r postgresLivestreamRepo) GetAllLivestreamings(page int) ([]domains.Livestream, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), `
		SELECT *
		FROM livestreams 
		WHERE ended_at IS NULL AND visibility = 'public'
		OFFSET $1 
		LIMIT $2
	`, 10*page, 10)
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

func (r postgresLivestreamRepo) GetPopularVODs(page int) ([]domains.Livestream, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), `
		SELECT *
		FROM livestreams 
		WHERE ended_at IS NOT NULL AND status = 'ended' AND visibility = 'public'
		ORDER BY view_count DESC
		OFFSET $1 LIMIT $2
	`, 10*page, 10)
	if err != nil {
		return nil, servererrors.ErrDatabaseQuery
	}

	livestreams, err := pgx.CollectRows(rows, pgx.RowToStructByName[domains.Livestream])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domains.Livestream{}, nil
		}

		return nil, servererrors.ErrDatabaseIssue
	}

	return livestreams, nil
}

func (r postgresLivestreamRepo) CheckIsUserLivestreaming(userId uuid.UUID) (bool, *servererrors.ServerError) {
	var exists bool
	err := r.dbConn.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1 
			FROM livestreams 
			WHERE ended_at IS NULL AND user_id = $1
		)
	`, userId).Scan(&exists)
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

	rows, err := r.dbConn.Query(context.Background(), `
		INSERT INTO livestreams (title, user_id, description, thumbnail_url, status, view_count, playback_url) 
		VALUES (@title, @user_id, @description, @thumbnail_url, @status, @view_count, @playback_url) 
		RETURNING *
	`, params)
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

func (r postgresLivestreamRepo) Update(livestream domains.Livestream) (*domains.Livestream, *servererrors.ServerError) {
	params := pgx.NamedArgs{
		"title":         livestream.Title,
		"description":   livestream.Description,
		"status":        livestream.Status,
		"view_count":    livestream.ViewCount,
		"visibility":    livestream.Visibility,
		"thumbnail_url": livestream.ThumbnailURL,
		"playback_url":  livestream.PlaybackURL,
		"ended_at":      livestream.EndedAt,
		"duration":      livestream.Duration,
		"id":            livestream.Id,
	}
	rows, err := r.dbConn.Query(
		context.Background(),
		`
		UPDATE livestreams 
		SET title = @title, description = @description, status = @status, view_count = @view_count, visibility = @visibility, thumbnail_url = @thumbnail_url, playback_url = @playback_url, ended_at = @ended_at, duration = @duration, updated_at = NOW()
		WHERE id = @id
		RETURNING *
	`, params)
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

func (r postgresLivestreamRepo) Delete(livestreamId uuid.UUID) *servererrors.ServerError {
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
