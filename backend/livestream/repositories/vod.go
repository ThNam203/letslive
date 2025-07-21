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

type postgresVODRepo struct {
	dbConn *pgxpool.Pool
}

func NewVODRepository(conn *pgxpool.Pool) domains.VODRepository {
	return &postgresVODRepo{
		dbConn: conn,
	}
}

func (r postgresVODRepo) GetById(ctx context.Context, id uuid.UUID) (*domains.VOD, *serviceresponse.ServiceErrorResponse) {
	query := `
        select id, livestream_id, user_id, title, description, thumbnail_url, visibility, view_count, duration, playback_url, created_at, updated_at
        from vods
        where id = $1
    `
	rows, err := r.dbConn.Query(ctx, query, id)
	if err != nil {
		logger.Errorf("db query error [getvodbyid: %v]", err)
		return nil, serviceresponse.ErrDatabaseQuery
	}

	vod, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serviceresponse.ErrVODNotFound
		}
		logger.Errorf("db scan error [getvodbyid: %v]", err)
		return nil, serviceresponse.ErrQueryScanFailed
	}
	return &vod, nil
}

func (r postgresVODRepo) GetPublicVODsByUser(ctx context.Context, userId uuid.UUID, page, limit int) ([]domains.VOD, *serviceresponse.ServiceErrorResponse) {
	offset := limit * page
	rows, err := r.dbConn.Query(context.Background(), `
		SELECT * 
		FROM vods
		WHERE user_id = $1 AND visibility = 'public'
		ORDER BY created_at DESC
		OFFSET $2
		LIMIT $3
	`, userId, offset, limit)
	if err != nil {
		logger.Errorf("db exec error [getpublicvodsbyuser id=%s: %v]", userId, err)
		return nil, serviceresponse.ErrDatabaseQuery
	}
	defer rows.Close()

	vods, err := pgx.CollectRows(rows, pgx.RowToStructByName[domains.VOD])
	if err != nil {
		logger.Errorf("db scan error [getpublicvodsbyuser: %v]", err)
		return nil, serviceresponse.ErrQueryScanFailed
	}

	return vods, nil
}

func (r *postgresVODRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) *serviceresponse.ServiceErrorResponse {
	query := `
        update vods
        set view_count = view_count + 1
        where id = $1
    `
	result, err := r.dbConn.Exec(ctx, query, id)
	if err != nil {
		logger.Errorf("db exec error [incrementvodviewcount id=%s: %v]", id, err)
		return serviceresponse.ErrVODUpdateFailed
	}

	if result.RowsAffected() == 0 {
		logger.Warnf("attempted to increment view count for non-existent vod id %s", id)
		return serviceresponse.ErrVODNotFound
	}

	return nil
}

func (r *postgresVODRepo) GetByUser(ctx context.Context, userId uuid.UUID, page int, limit int) ([]domains.VOD, *serviceresponse.ServiceErrorResponse) {
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
		logger.Errorf("db query error [getvodbyuser: %v]", err)
		return nil, serviceresponse.ErrDatabaseQuery
	}

	vods, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		logger.Errorf("db scan error [getvodbyuser: %v]", err)
		return nil, serviceresponse.ErrQueryScanFailed
	}
	return vods, nil
}

func (r *postgresVODRepo) GetPopular(ctx context.Context, page int, limit int) ([]domains.VOD, *serviceresponse.ServiceErrorResponse) {
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
		logger.Errorf("db query error [getpopularvods: %v]", err)
		return nil, serviceresponse.ErrDatabaseQuery
	}

	vods, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		logger.Errorf("db scan error [getpopularvods: %v]", err)
		return nil, serviceresponse.ErrQueryScanFailed
	}
	return vods, nil
}

func (r *postgresVODRepo) Create(ctx context.Context, vod domains.VOD) (*domains.VOD, *serviceresponse.ServiceErrorResponse) {
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
		logger.Errorf("db query error [createvod: %v]", err)
		return nil, serviceresponse.ErrVODCreateFailed
	}

	createdVod, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		logger.Errorf("db scan error [createvod: %v]", err)
		return nil, serviceresponse.ErrQueryScanFailed
	}
	return &createdVod, nil
}

func (r *postgresVODRepo) Update(ctx context.Context, vod domains.VOD) (*domains.VOD, *serviceresponse.ServiceErrorResponse) {
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
		logger.Errorf("db query error [updatevod id=%s: %v]", vod.Id, err)
		return nil, serviceresponse.ErrVODUpdateFailed
	}

	updatedVod, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.VOD])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serviceresponse.ErrVODNotFound
		}
		logger.Errorf("db scan error [updatevod id=%s: %v]", vod.Id, err)
		return nil, serviceresponse.ErrQueryScanFailed
	}
	return &updatedVod, nil
}

func (r *postgresVODRepo) Delete(ctx context.Context, id uuid.UUID) *serviceresponse.ServiceErrorResponse {
	result, err := r.dbConn.Exec(ctx, "delete from vods where id = $1", id)
	if err != nil {
		logger.Errorf("db exec error [deletevod id=%s: %v]", id, err)
		return serviceresponse.ErrDatabaseQuery
	}
	if result.RowsAffected() == 0 {
		return serviceresponse.ErrVODNotFound
	}
	return nil
}
