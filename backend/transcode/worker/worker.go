package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/transcode/config"
	"sen1or/letslive/transcode/domains"
	"sen1or/letslive/transcode/storage"
	"sen1or/letslive/transcode/upload"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	miniocreds "github.com/minio/minio-go/v7/pkg/credentials"
)

type LivestreamGateway interface {
	UpdateVODStatus(ctx context.Context, vodId string, status domains.VODStatus, playbackUrl string, thumbnailUrl string, duration *int64) error
}

type TranscodeWorker struct {
	db                *pgxpool.Pool
	hlsStorage        storage.Storage
	rawMinioClient    *minio.Client
	rawMinioBucket    string
	config            *config.Config
	livestreamGateway LivestreamGateway
	stopChan          chan struct{}
}

func NewTranscodeWorker(
	db *pgxpool.Pool,
	hlsStorage storage.Storage,
	rawMinioClient *minio.Client,
	rawMinioBucket string,
	cfg *config.Config,
	livestreamGateway LivestreamGateway,
) *TranscodeWorker {
	return &TranscodeWorker{
		db:                db,
		hlsStorage:        hlsStorage,
		rawMinioClient:    rawMinioClient,
		rawMinioBucket:    rawMinioBucket,
		config:            cfg,
		livestreamGateway: livestreamGateway,
		stopChan:          make(chan struct{}),
	}
}

func NewRawMinIOClient(cfg config.MinIO) *minio.Client {
	client, err := minio.New(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), &minio.Options{
		Creds:  miniocreds.NewStaticV4(os.Getenv("MINIO_ROOT_USER"), os.Getenv("MINIO_ROOT_PASSWORD"), ""),
		Secure: false,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create raw MinIO client: %v", err))
	}
	return client
}

func (w *TranscodeWorker) Start(ctx context.Context) {
	logger.Infof(ctx, "transcode worker started, polling every 5 seconds")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Infof(ctx, "transcode worker stopping (context cancelled)")
			return
		case <-w.stopChan:
			logger.Infof(ctx, "transcode worker stopping (stop signal)")
			return
		case <-ticker.C:
			w.processNextJob(ctx)
		}
	}
}

func (w *TranscodeWorker) Shutdown() {
	close(w.stopChan)
}

func (w *TranscodeWorker) processNextJob(ctx context.Context) {
	tx, err := w.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Errorf(ctx, "worker: failed to begin tx: %v", err)
		return
	}
	defer tx.Rollback(ctx)

	// Poll for a pending job with row-level locking
	var jobId, vodId string
	var attempts, maxAttempts int
	var originalFileURL, prevError *string

	err = tx.QueryRow(ctx, `
		SELECT tj.id, tj.vod_id, tj.attempts, tj.max_attempts, v.original_file_url, tj.error_message
		FROM transcode_jobs tj
		JOIN vods v ON v.id = tj.vod_id
		WHERE tj.status = 'pending' AND tj.attempts < tj.max_attempts
		ORDER BY tj.created_at ASC
		LIMIT 1
		FOR UPDATE OF tj SKIP LOCKED
	`).Scan(&jobId, &vodId, &attempts, &maxAttempts, &originalFileURL, &prevError)

	if err != nil {
		if err == pgx.ErrNoRows {
			tx.Rollback(ctx)
			return // no pending jobs
		}
		logger.Errorf(ctx, "worker: failed to query pending job: %v", err)
		return
	}

	if originalFileURL == nil || *originalFileURL == "" {
		logger.Errorf(ctx, "worker: vod %s has no original file URL", vodId)
		w.failJob(ctx, tx, jobId, vodId, "no original file URL", attempts, maxAttempts)
		return
	}

	// Mark job as processing
	_, err = tx.Exec(ctx, `UPDATE transcode_jobs SET status = 'processing', started_at = now(), attempts = attempts + 1, updated_at = now() WHERE id = $1`, jobId)
	if err != nil {
		logger.Errorf(ctx, "worker: failed to mark job processing: %v", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Errorf(ctx, "worker: failed to commit tx: %v", err)
		return
	}

	logger.Infof(ctx, "worker: processing job %s for vod %s", jobId, vodId)

	// Do the actual transcoding work
	if err := w.doTranscode(ctx, jobId, vodId, *originalFileURL, attempts+1, maxAttempts, forceConvert(prevError)); err != nil {
		logger.Errorf(ctx, "worker: transcode failed for job %s: %v", jobId, err)
	}
}

func (w *TranscodeWorker) doTranscode(ctx context.Context, jobId, vodId, rawFileURL string, currentAttempt, maxAttempts int, convertAll bool) error {
	// Extract the object path from the raw file URL
	// URL format: http://host:port/bucket/raw-videos/vodId/filename
	objectName := extractObjectName(rawFileURL, w.rawMinioBucket)
	if objectName == "" {
		errMsg := "failed to extract object name from URL"
		w.markJobFailed(ctx, jobId, vodId, errMsg, currentAttempt, maxAttempts)
		return errors.New(errMsg)
	}

	// Create temp directory for this job
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("transcode-%s-*", vodId))
	if err != nil {
		errMsg := fmt.Sprintf("failed to create temp dir: %v", err)
		w.markJobFailed(ctx, jobId, vodId, errMsg, currentAttempt, maxAttempts)
		return errors.New(errMsg)
	}
	defer os.RemoveAll(tempDir)

	// Download raw file from MinIO
	rawFilePath := filepath.Join(tempDir, "input"+filepath.Ext(objectName))
	if err := w.downloadFromMinIO(ctx, objectName, rawFilePath); err != nil {
		errMsg := fmt.Sprintf("failed to download raw file: %v", err)
		w.markJobFailed(ctx, jobId, vodId, errMsg, currentAttempt, maxAttempts)
		return errors.New(errMsg)
	}

	// Run the file through the live engine, as a recorded livestream
	if err := prepareAttempt(w.config.Transcode.PublicHLSPath, vodId); err != nil {
		errMsg := fmt.Sprintf("failed to clear the previous attempt: %v", err)
		w.markJobFailed(ctx, jobId, vodId, errMsg, currentAttempt, maxAttempts)
		return errors.New(errMsg)
	}
	if convertAll {
		logger.Infof(ctx, "worker: vod %s: the previous attempt failed on the input, converting every stream", vodId)
	}
	res, err := upload.Run(ctx, w.config.Transcode, w.hlsStorage, vodId, rawFilePath, convertAll)
	if err != nil {
		errMsg := failureMessage(err)
		w.markJobFailed(ctx, jobId, vodId, errMsg, currentAttempt, maxAttempts)
		return errors.New(errMsg)
	}
	playbackURL := fmt.Sprintf("%s/%s/%s", w.config.Transcode.VODPlaybackUrlPrefix, vodId, w.config.Transcode.FFMpegSetting.MasterFileName)
	duration := int64(math.Ceil(res.Duration.Seconds()))

	// Update VOD status to ready via livestream gateway
	if err := w.livestreamGateway.UpdateVODStatus(ctx, vodId, domains.VODStatusReady, playbackURL, res.ThumbnailURL, &duration); err != nil {
		errMsg := fmt.Sprintf("failed to update VOD status: %v", err)
		w.markJobFailed(ctx, jobId, vodId, errMsg, currentAttempt, maxAttempts)
		return errors.New(errMsg)
	}

	// Mark job completed
	_, err = w.db.Exec(ctx, `UPDATE transcode_jobs SET status = 'completed', completed_at = now(), updated_at = now() WHERE id = $1`, jobId)
	if err != nil {
		logger.Errorf(ctx, "worker: failed to mark job completed: %v", err)
	}

	// Clean up raw file from MinIO
	if err := w.rawMinioClient.RemoveObject(ctx, w.rawMinioBucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		logger.Warnf(ctx, "worker: failed to delete raw file %s: %v", objectName, err)
	}

	logger.Infof(ctx, "worker: job %s completed successfully for vod %s", jobId, vodId)
	return nil
}

func (w *TranscodeWorker) downloadFromMinIO(ctx context.Context, objectName, destPath string) error {
	obj, err := w.rawMinioClient.GetObject(ctx, w.rawMinioBucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object from minio: %w", err)
	}
	defer obj.Close()

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create dest file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, obj); err != nil {
		return fmt.Errorf("failed to copy object to file: %w", err)
	}

	return nil
}

func (w *TranscodeWorker) markJobFailed(ctx context.Context, jobId, vodId, errMsg string, currentAttempt, maxAttempts int) {
	if currentAttempt >= maxAttempts {
		// Max attempts reached, mark as failed permanently
		_, err := w.db.Exec(ctx, `UPDATE transcode_jobs SET status = 'failed', error_message = $1, updated_at = now() WHERE id = $2`, errMsg, jobId)
		if err != nil {
			logger.Errorf(ctx, "worker: failed to mark job as failed: %v", err)
		}
		// Mark VOD as failed too
		w.livestreamGateway.UpdateVODStatus(ctx, vodId, domains.VODStatusFailed, "", "", nil)
	} else {
		// Reset to pending for retry
		_, err := w.db.Exec(ctx, `UPDATE transcode_jobs SET status = 'pending', error_message = $1, updated_at = now() WHERE id = $2`, errMsg, jobId)
		if err != nil {
			logger.Errorf(ctx, "worker: failed to reset job to pending: %v", err)
		}
	}
}

func (w *TranscodeWorker) failJob(ctx context.Context, tx pgx.Tx, jobId, vodId, errMsg string, attempts, maxAttempts int) {
	_, err := tx.Exec(ctx, `UPDATE transcode_jobs SET status = 'failed', error_message = $1, attempts = attempts + 1, updated_at = now() WHERE id = $2`, errMsg, jobId)
	if err != nil {
		logger.Errorf(ctx, "worker: failed to mark job as failed: %v", err)
	}
	tx.Commit(ctx)
	w.livestreamGateway.UpdateVODStatus(ctx, vodId, domains.VODStatusFailed, "", "", nil)
}

// inputErrPrefix starts the stored error of an attempt that failed on the
// input file; the next attempt then converts every stream.
const inputErrPrefix = "input: "

// failureMessage is the error stored for a failed upload attempt.
func failureMessage(err error) string {
	if errors.Is(err, upload.ErrInput) {
		return inputErrPrefix + err.Error()
	}
	return "transcode failed: " + err.Error()
}

// forceConvert reports whether the previous attempt failed on the input, so
// this one must convert every stream instead of copying.
func forceConvert(prevError *string) bool {
	return prevError != nil && strings.HasPrefix(*prevError, inputErrPrefix)
}

// prepareAttempt removes the playlists a failed earlier attempt of vodId may
// have left, so a retry starts clean.
func prepareAttempt(publicHLSPath, vodId string) error {
	if vodId == "" || vodId == "." || vodId == ".." || strings.ContainsAny(vodId, `/\`) {
		return fmt.Errorf("unsafe vod id %q", vodId)
	}
	return os.RemoveAll(filepath.Join(publicHLSPath, "vods", vodId))
}

// extractObjectName extracts the MinIO object path from a full URL.
// URL format: http://host:port/bucket/object/path
func extractObjectName(rawURL, bucket string) string {
	idx := strings.Index(rawURL, "/"+bucket+"/")
	if idx == -1 {
		return ""
	}
	return rawURL[idx+len("/"+bucket+"/"):]
}
