package repositories

import (
	"context"
	"errors"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/pkg/logger"
	serviceresponse "sen1or/letslive/livestream/responses"

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

func (r *postgresLivestreamRepo) GetById(ctx context.Context, id uuid.UUID) (*domains.Livestream, *serviceresponse.ServiceErrorResponse) {
	query := `
		SELECT id, user_id, title, description, thumbnail_url, visibility, view_count, started_at, ended_at, created_at, updated_at, vod_id
		FROM livestreams
		WHERE id = $1
	`

	rows, err := r.dbConn.Query(ctx, query, id)
	if err != nil {
		logger.Errorf("db query error [getlivestreambyid: %v]", err)
		return nil, serviceresponse.ErrDatabaseQuery
	}

	livestream, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serviceresponse.ErrLivestreamNotFound
		}
		logger.Errorf("db scan error [getlivestreambyid: %v]", err)
		return nil, serviceresponse.ErrQueryScanFailed
	}
	return &livestream, nil
}

// TODO: implement a recommendation system
func (r *postgresLivestreamRepo) GetRecommendedLivestreams(ctx context.Context, page int, limit int) ([]domains.Livestream, *serviceresponse.ServiceErrorResponse) {
	offset := limit * page
	query := `
		SELECT id, user_id, title, description, thumbnail_url, visibility, view_count, started_at, ended_at, created_at, updated_at, vod_id
		FROM livestreams
        	WHERE ended_at IS NULL AND visibility = 'public'
        	ORDER BY started_at DESC
        	OFFSET $1 LIMIT $2
	`
	rows, err := r.dbConn.Query(ctx, query, offset, limit)
	if err != nil {
		logger.Errorf("db query error [getalllivestreamings: %v]", err)
		return nil, serviceresponse.ErrDatabaseQuery
	}

	livestreams, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		logger.Errorf("db scan error [getalllivestreamings: %v]", err)
		return nil, serviceresponse.ErrQueryScanFailed
	}

	return livestreams, nil
}
func (r *postgresLivestreamRepo) CheckIsUserLivestreaming(ctx context.Context, userId uuid.UUID) (bool, *serviceresponse.ServiceErrorResponse) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM livestreams WHERE ended_at IS NULL AND user_id = $1)`
	err := r.dbConn.QueryRow(ctx, query, userId).Scan(&exists)
	if err != nil {
		logger.Errorf("db scan error [checkisuserlivestreaming: %v]", err)
		return false, serviceresponse.ErrQueryScanFailed
	}

	return exists, nil
}

func (r *postgresLivestreamRepo) Create(ctx context.Context, newLivestream domains.Livestream) (*domains.Livestream, *serviceresponse.ServiceErrorResponse) {
	query := `
		INSERT INTO livestreams (user_id, title, description, thumbnail_url, visibility)
        	VALUES ($1, $2, $3, $4, $5)
        	RETURNING id, user_id, title, description, thumbnail_url, visibility, view_count, started_at, ended_at, created_at, updated_at, vod_id
	`
	rows, err := r.dbConn.Query(ctx, query,
		newLivestream.UserId,
		newLivestream.Title,
		newLivestream.Description,
		newLivestream.ThumbnailURL,
		newLivestream.Visibility,
	)
	if err != nil {
		logger.Errorf("db query error [createlivestream: %v]", err)
		return nil, serviceresponse.ErrLivestreamCreateFailed
	}

	createdLs, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		logger.Errorf("db scan error [createlivestream: %v]", err)
		return nil, serviceresponse.ErrQueryScanFailed
	}
	return &createdLs, nil
}

func (r *postgresLivestreamRepo) Update(ctx context.Context, livestream domains.Livestream) (*domains.Livestream, *serviceresponse.ServiceErrorResponse) {
	query := `
		UPDATE livestreams
		SET title = $1, description = $2, thumbnail_url = $3, visibility = $4, ended_at = $5, vod_id = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING id, user_id, title, description, thumbnail_url, visibility, view_count, started_at, ended_at, created_at, updated_at, vod_id
	`

	rows, err := r.dbConn.Query(ctx, query,
		livestream.Title,
		livestream.Description,
		livestream.ThumbnailURL,
		livestream.Visibility,
		livestream.EndedAt,
		livestream.VODId,
		livestream.Id,
	)
	if err != nil {
		logger.Errorf("db query error [updatelivestream id=%s: %v]", livestream.Id, err)
		return nil, serviceresponse.ErrLivestreamUpdateFailed
	}

	updatedLs, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serviceresponse.ErrLivestreamNotFound
		}
		logger.Errorf("db scan error [updatelivestream id=%s: %v]", livestream.Id, err)
		return nil, serviceresponse.ErrQueryScanFailed
	}
	return &updatedLs, nil
}

func (r *postgresLivestreamRepo) Delete(ctx context.Context, livestreamId uuid.UUID) *serviceresponse.ServiceErrorResponse {
	result, err := r.dbConn.Exec(ctx, `
		DELETE FROM livestreams 
		WHERE id = $1
	`, livestreamId)

	if err != nil {
		logger.Errorf("db exec error [deletelivestream id=%s: %v]", livestreamId, err)
		return serviceresponse.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return serviceresponse.ErrLivestreamNotFound
	}
	return nil
}
