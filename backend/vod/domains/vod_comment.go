package domains

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX is an interface satisfied by both *pgxpool.Pool and pgx.Tx,
// allowing repository methods to work with either.
type DBTX interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type VODComment struct {
	Id         uuid.UUID  `json:"id" db:"id"`
	VODId      uuid.UUID  `json:"vodId" db:"vod_id"`
	UserId     uuid.UUID  `json:"userId" db:"user_id"`
	ParentId   *uuid.UUID `json:"parentId" db:"parent_id"`
	Content    string     `json:"content" db:"content"`
	IsDeleted  bool       `json:"isDeleted" db:"is_deleted"`
	IsEdited   bool       `json:"isEdited" db:"is_edited"`
	LikeCount  int64      `json:"likeCount" db:"like_count"`
	ReplyCount int64      `json:"replyCount" db:"reply_count"`
	CreatedAt  time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time  `json:"updatedAt" db:"updated_at"`
}

type VODCommentLike struct {
	Id        uuid.UUID `json:"id" db:"id"`
	CommentId uuid.UUID `json:"commentId" db:"comment_id"`
	UserId    uuid.UUID `json:"userId" db:"user_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type VODCommentRepository interface {
	WithTx(tx pgx.Tx) VODCommentRepository
	GetByVODId(ctx context.Context, vodId uuid.UUID, page int, limit int) ([]VODComment, error)
	CountByVODId(ctx context.Context, vodId uuid.UUID) (int, error)
	CountReplies(ctx context.Context, parentId uuid.UUID) (int, error)
	GetReplies(ctx context.Context, parentId uuid.UUID, page int, limit int) ([]VODComment, error)
	GetById(ctx context.Context, id uuid.UUID) (*VODComment, error)
	Create(ctx context.Context, comment VODComment) (*VODComment, error)
	UpdateContent(ctx context.Context, id uuid.UUID, content string) (*VODComment, error)
	IncrementReplyCount(ctx context.Context, commentId uuid.UUID) error
	DecrementReplyCount(ctx context.Context, commentId uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

// VODCommentEdit is the content a comment held before one specific edit
// replaced it. The current content is never stored here: it stays on the
// comment itself, as the latest version.
type VODCommentEdit struct {
	Id              uuid.UUID `json:"id" db:"id"`
	CommentId       uuid.UUID `json:"commentId" db:"comment_id"`
	Version         int       `json:"version" db:"version"`
	PreviousContent string    `json:"previousContent" db:"previous_content"`
	EditedAt        time.Time `json:"editedAt" db:"edited_at"`
}

type VODCommentEditRepository interface {
	WithTx(tx pgx.Tx) VODCommentEditRepository
	// Insert records previousContent as the next version of the comment's
	// history. It returns ErrCommentUpdateFailed if another edit already
	// claimed that version, which keeps concurrent edits from overwriting
	// each other's history entries.
	Insert(ctx context.Context, commentId uuid.UUID, previousContent string) error
}

type VODCommentLikeRepository interface {
	WithTx(tx pgx.Tx) VODCommentLikeRepository
	GetUserLikedCommentIds(ctx context.Context, commentIds []uuid.UUID, userId uuid.UUID) ([]uuid.UUID, error)
	InsertLike(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error
	DeleteLike(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) error
	IncrementLikeCount(ctx context.Context, commentId uuid.UUID) error
	DecrementLikeCount(ctx context.Context, commentId uuid.UUID) error
}
