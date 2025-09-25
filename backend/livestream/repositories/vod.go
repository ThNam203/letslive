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

type postgresVODRepo struct {
	dbConn *pgxpool.Pool
}

func NewVODRepository(conn *pgxpool.Pool) domains.VODRepository {
	return &postgresVODRepo{
		dbConn: conn,
	}
}

func (r postgresVODRepo) GetById(ctx context.Context, id uuid.UUID) (*domains.VOD, *response.Response[any]) {
	query := `
        select id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, duration, playback_url, created_at, updated_at
        from vods
        where id = $1
    `
	rows, err := r.dbConn.Query(ctx, query, id)
	if err != nil {
		logger.Errorf(ctx, "db query error [getvodbyid: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	vod, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_VOD_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}
		logger.Errorf(ctx, "db scan error [getvodbyid: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &vod, nil
}

func (r postgresVODRepo) GetPublicVODsByUser(ctx context.Context, userId uuid.UUID, page, limit int) ([]domains.VOD, *response.Response[any]) {
	offset := limit * page
	rows, err := r.dbConn.Query(ctx, `
		SELECT * 
		FROM vods
		WHERE user_id = $1 AND visibility = 'public'
		ORDER BY created_at DESC
		OFFSET $2
		LIMIT $3
	`, userId, offset, limit)
	if err != nil {
		logger.Errorf(ctx, "db exec error [getpublicvodsbyuser id=%s: %v]", userId, err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	defer rows.Close()

	vods, err := pgx.CollectRows(rows, pgx.RowToStructByName[domains.VOD])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getpublicvodsbyuser: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return vods, nil
}

func (r *postgresVODRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) *response.Response[any] {
	query := `
        update vods
        set view_count = view_count + 1
        where id = $1
    `
	result, err := r.dbConn.Exec(ctx, query, id)
	if err != nil {
		logger.Errorf(ctx, "db exec error [incrementvodviewcount id=%s: %v]", id, err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_UPDATE_FAILED,
			nil,
			nil,
			nil,
		)
	}

	if result.RowsAffected() == 0 {
		logger.Warnf(ctx, "attempted to increment view count for non-existent vod id %s", id)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}

	return nil
}

func (r *postgresVODRepo) GetByUser(ctx context.Context, userId uuid.UUID, page int, limit int) ([]domains.VOD, *response.Response[any]) {
	offset := limit * page
	query := `
        select id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, duration, playback_url, created_at, updated_at
        from vods
        where user_id = $1
        order by created_at desc
        offset $2 limit $3
    `
	rows, err := r.dbConn.Query(ctx, query, userId, offset, limit)
	if err != nil {
		logger.Errorf(ctx, "db query error [getvodbyuser: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	vods, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getvodbyuser: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return vods, nil
}

func (r *postgresVODRepo) GetPopular(ctx context.Context, page int, limit int) ([]domains.VOD, *response.Response[any]) {
	offset := limit * page
	query := `
        select id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, duration, playback_url, created_at, updated_at
        from vods
        where visibility = 'public'
        order by view_count desc
        offset $1 limit $2
    `
	rows, err := r.dbConn.Query(ctx, query, offset, limit)
	if err != nil {
		logger.Errorf(ctx, "db query error [getpopularvods: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	vods, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		logger.Errorf(ctx, "db scan error [getpopularvods: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return vods, nil
}

func (r *postgresVODRepo) Create(ctx context.Context, vod domains.VOD) (*domains.VOD, *response.Response[any]) {
	query := `
        insert into vods (livestream_id, user_id, title, description, thumbnail_url, visibility, duration, playback_url, view_count, created_at)
        values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        returning id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, duration, playback_url, created_at, updated_at
    `
	rows, err := r.dbConn.Query(ctx, query,
		vod.LivestreamId, vod.UserId, vod.Title, vod.Description, vod.ThumbnailURL,
		vod.Visibility, vod.Duration, vod.PlaybackURL, vod.ViewCount, vod.CreatedAt,
	)

	if err != nil {
		// todo: check for specific db errors like fk violations if possible
		logger.Errorf(ctx, "db query error [createvod: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_CREATE_FAILED,
			nil,
			nil,
			nil,
		)
	}

	createdVod, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		logger.Errorf(ctx, "db scan error [createvod: %v]", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &createdVod, nil
}

func (r *postgresVODRepo) Update(ctx context.Context, vod domains.VOD) (*domains.VOD, *response.Response[any]) {
	query := `
        update vods
        set title = $1, description = $2, thumbnail_url = $3, visibility = $4, duration = $5, playback_url = $6, updated_at = now()
        where id = $7
        returning id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, duration, playback_url, created_at, updated_at
    `
	rows, err := r.dbConn.Query(ctx, query,
		vod.Title, vod.Description, vod.ThumbnailURL, vod.Visibility,
		vod.Duration, vod.PlaybackURL, vod.Id,
	)
	if err != nil {
		logger.Errorf(ctx, "db query error [updatevod id=%s: %v]", vod.Id, err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_UPDATE_FAILED,
			nil,
			nil,
			nil,
		)
	}

	updatedVod, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_VOD_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}
		logger.Errorf(ctx, "db scan error [updatevod id=%s: %v]", vod.Id, err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}
	return &updatedVod, nil
}

func (r *postgresVODRepo) Delete(ctx context.Context, id uuid.UUID) *response.Response[any] {
	result, err := r.dbConn.Exec(ctx, "delete from vods where id = $1", id)
	if err != nil {
		logger.Errorf(ctx, "db exec error [deletevod id=%s: %v]", id, err)
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	if result.RowsAffected() == 0 {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_VOD_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}
	return nil
}
