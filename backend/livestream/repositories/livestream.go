package repositories

import (
	"context"
	"errors"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/pkg/logger"
	"sen1or/letslive/livestream/response"

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

func (r *postgresLivestreamRepo) GetById(ctx context.Context, id uuid.UUID) (*domains.Livestream, *response.Response[any]) {
	query := `
		SELECT id, user_id, title, description, thumbnail_url, visibility, view_count, started_at, ended_at, created_at, updated_at, vod_id
		FROM livestreams
		WHERE id = $1
	`

	rows, err := r.dbConn.Query(ctx, query, id)
	if err != nil {
		logger.Errorf(ctx, "db query error [getlivestreambyid: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	livestream, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_LIVESTREAM_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}
		logger.Errorf(ctx, "db scan error [getlivestreambyid: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &livestream, nil
}

func (r *postgresLivestreamRepo) GetByUser(ctx context.Context, userId uuid.UUID) (*domains.Livestream, *response.Response[any]) {
	query := `
		SELECT id, user_id, title, description, thumbnail_url, visibility, view_count, started_at, ended_at, created_at, updated_at, vod_id
		FROM livestreams
		WHERE user_id = $1 AND vod_id IS NULL
		LIMIT 1
	`

	rows, err := r.dbConn.Query(ctx, query, userId)
	if err != nil {
		logger.Errorf(ctx, "db query error [getlivestreambyuser: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	livestream, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_LIVESTREAM_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}
		logger.Errorf(ctx, "db scan error [getlivestreambyuser: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &livestream, nil
}

// TODO: implement a recommendation system
func (r *postgresLivestreamRepo) GetRecommendedLivestreams(ctx context.Context, page int, limit int) ([]domains.Livestream, *response.Response[any]) {
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
		logger.Errorf(ctx, "db query error [getalllivestreamings: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	livestreams, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getalllivestreamings: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return livestreams, nil
}

func (r *postgresLivestreamRepo) Create(ctx context.Context, newLivestream domains.Livestream) (*domains.Livestream, *response.Response[any]) {
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
		logger.Errorf(ctx, "db query error [createlivestream: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_LIVESTREAM_CREATE_FAILED,
			nil,
			nil,
			nil,
		)
	}

	createdLs, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		logger.Errorf(ctx, "db scan error [createlivestream: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &createdLs, nil
}

func (r *postgresLivestreamRepo) Update(ctx context.Context, livestream domains.Livestream) (*domains.Livestream, *response.Response[any]) {
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
		logger.Errorf(ctx, "db query error [updatelivestream id=%s: %v]", livestream.Id, err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_LIVESTREAM_UPDATE_FAILED,
			nil,
			nil,
			nil,
		)
	}

	updatedLs, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.Livestream])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_LIVESTREAM_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}
		logger.Errorf(ctx, "db scan error [updatelivestream id=%s: %v]", livestream.Id, err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &updatedLs, nil
}

func (r *postgresLivestreamRepo) Delete(ctx context.Context, livestreamId uuid.UUID) *response.Response[any] {
	result, err := r.dbConn.Exec(ctx, `
		DELETE FROM livestreams 
		WHERE id = $1
	`, livestreamId)
	if err != nil {
		logger.Errorf(ctx, "db exec error [deletelivestream id=%s: %v]", livestreamId, err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	if result.RowsAffected() == 0 {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_LIVESTREAM_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}
	return nil
}
