package domains

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

type VODVisibility string

const (
	VODPublicVisibility  VODVisibility = "public"
	VODPrivateVisibility VODVisibility = "private"
)

type VODStatus string

const (
	VODStatusUploading  VODStatus = "uploading"
	VODStatusProcessing VODStatus = "processing"
	VODStatusReady      VODStatus = "ready"
	VODStatusFailed     VODStatus = "failed"
)

type VOD struct {
	Id              uuid.UUID     `json:"id" db:"id"`
	LivestreamId    *uuid.UUID    `json:"livestreamId" db:"livestream_id"`
	UserId          uuid.UUID     `json:"userId" db:"user_id"`
	Title           string        `json:"title" db:"title"`
	Description     *string       `json:"description" db:"description"`
	ThumbnailURL    *string       `json:"thumbnailUrl" db:"thumbnail_url"`
	Visibility      VODVisibility `json:"visibility" db:"visibility"`
	ViewCount       int64         `json:"viewCount" db:"view_count"`
	Duration        int64         `json:"duration" db:"duration"`
	PlaybackURL     *string       `json:"playbackUrl" db:"playback_url"`
	Status          VODStatus     `json:"status" db:"status"`
	OriginalFileURL *string       `json:"originalFileUrl" db:"original_file_url"`
	CreatedAt       time.Time     `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time     `json:"updatedAt" db:"updated_at"`
}

type TranscodeJobStatus string

const (
	TranscodeJobPending    TranscodeJobStatus = "pending"
	TranscodeJobProcessing TranscodeJobStatus = "processing"
	TranscodeJobCompleted  TranscodeJobStatus = "completed"
	TranscodeJobFailed     TranscodeJobStatus = "failed"
)

type TranscodeJob struct {
	Id          uuid.UUID          `json:"id" db:"id"`
	VodId       uuid.UUID          `json:"vodId" db:"vod_id"`
	Status      TranscodeJobStatus `json:"status" db:"status"`
	Attempts    int                `json:"attempts" db:"attempts"`
	MaxAttempts int                `json:"maxAttempts" db:"max_attempts"`
	ErrorMsg    *string            `json:"errorMessage,omitempty" db:"error_message"`
	CreatedAt   time.Time          `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time          `json:"updatedAt" db:"updated_at"`
	StartedAt   *time.Time         `json:"startedAt,omitempty" db:"started_at"`
	CompletedAt *time.Time         `json:"completedAt,omitempty" db:"completed_at"`
}

type VODRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*VOD, error)
	GetByUser(ctx context.Context, userId uuid.UUID, page int, limit int) ([]VOD, error)
	GetPublicVODsByUser(ctx context.Context, userId uuid.UUID, page int, limit int) ([]VOD, error)
	GetPopular(ctx context.Context, page int, limit int) ([]VOD, int, error)
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	Create(ctx context.Context, vod VOD) (*VOD, error)
	Update(ctx context.Context, vod VOD) (*VOD, error)
	UpdateStatus(ctx context.Context, vodId uuid.UUID, status VODStatus, playbackUrl *string, thumbnailUrl *string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TranscodeJobRepository interface {
	Create(ctx context.Context, job TranscodeJob) (*TranscodeJob, error)
	GetPendingJob(ctx context.Context) (*TranscodeJob, error)
	UpdateStatus(ctx context.Context, jobId uuid.UUID, status TranscodeJobStatus, errorMsg *string) error
	IncrementAttempts(ctx context.Context, jobId uuid.UUID) error
}
