package transcodejob

import (
	"context"
	"errors"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresTranscodeJobRepo struct {
	dbConn *pgxpool.Pool
}

func NewTranscodeJobRepository(conn *pgxpool.Pool) domains.TranscodeJobRepository {
	return &postgresTranscodeJobRepo{
		dbConn: conn,
	}
}

func (r *postgresTranscodeJobRepo) Create(ctx context.Context, job domains.TranscodeJob) (*domains.TranscodeJob, error) {
	query := `
		INSERT INTO transcode_jobs (vod_id, status, attempts, max_attempts)
		VALUES ($1, $2, $3, $4)
		RETURNING id, vod_id, status, attempts, max_attempts, error_message, created_at, updated_at, started_at, completed_at
	`
	rows, err := r.dbConn.Query(ctx, query, job.VodId, job.Status, job.Attempts, job.MaxAttempts)
	if err != nil {
		logger.Errorf(ctx, "db query error [create_transcode_job: %v]", err)
		return nil, domains.ErrInternal
	}

	createdJob, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.TranscodeJob])
	if err != nil {
		logger.Errorf(ctx, "db scan error [create_transcode_job: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &createdJob, nil
}

func (r *postgresTranscodeJobRepo) GetPendingJob(ctx context.Context) (*domains.TranscodeJob, error) {
	query := `
		SELECT id, vod_id, status, attempts, max_attempts, error_message, created_at, updated_at, started_at, completed_at
		FROM transcode_jobs
		WHERE status = 'pending' AND attempts < max_attempts
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`
	rows, err := r.dbConn.Query(ctx, query)
	if err != nil {
		logger.Errorf(ctx, "db query error [get_pending_job: %v]", err)
		return nil, domains.ErrDatabaseQuery
	}

	job, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.TranscodeJob])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // no pending jobs
		}
		logger.Errorf(ctx, "db scan error [get_pending_job: %v]", err)
		return nil, domains.ErrDatabaseIssue
	}
	return &job, nil
}

func (r *postgresTranscodeJobRepo) UpdateStatus(ctx context.Context, jobId uuid.UUID, status domains.TranscodeJobStatus, errorMsg *string) error {
	var query string
	var err error

	now := time.Now()

	switch status {
	case domains.TranscodeJobProcessing:
		query = `UPDATE transcode_jobs SET status = $1, started_at = $2, updated_at = $2 WHERE id = $3`
		_, err = r.dbConn.Exec(ctx, query, status, now, jobId)
	case domains.TranscodeJobCompleted:
		query = `UPDATE transcode_jobs SET status = $1, completed_at = $2, updated_at = $2 WHERE id = $3`
		_, err = r.dbConn.Exec(ctx, query, status, now, jobId)
	case domains.TranscodeJobFailed:
		query = `UPDATE transcode_jobs SET status = $1, error_message = $2, updated_at = $3 WHERE id = $4`
		_, err = r.dbConn.Exec(ctx, query, status, errorMsg, now, jobId)
	default:
		query = `UPDATE transcode_jobs SET status = $1, updated_at = $2 WHERE id = $3`
		_, err = r.dbConn.Exec(ctx, query, status, now, jobId)
	}

	if err != nil {
		logger.Errorf(ctx, "db query error [update_job_status id=%s: %v]", jobId, err)
		return domains.ErrInternal
	}

	return nil
}

func (r *postgresTranscodeJobRepo) IncrementAttempts(ctx context.Context, jobId uuid.UUID) error {
	query := `UPDATE transcode_jobs SET attempts = attempts + 1, updated_at = now() WHERE id = $1`
	_, err := r.dbConn.Exec(ctx, query, jobId)
	if err != nil {
		logger.Errorf(ctx, "db query error [increment_attempts id=%s: %v]", jobId, err)
		return domains.ErrInternal
	}
	return nil
}
