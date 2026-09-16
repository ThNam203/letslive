package domains

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

const NotificationTypeGiftReceived = "gift_received"

type Notification struct {
	Id          uuid.UUID  `json:"id" db:"id"`
	UserId      uuid.UUID  `json:"userId" db:"user_id"`
	Type        string     `json:"type" db:"type"`
	Title       string     `json:"title" db:"title"`
	Message     string     `json:"message" db:"message"`
	ActionUrl   *string    `json:"actionUrl" db:"action_url"`
	ActionLabel *string    `json:"actionLabel" db:"action_label"`
	ReferenceId *uuid.UUID `json:"referenceId" db:"reference_id"`
	IsRead      bool       `json:"isRead" db:"is_read"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
}

type NotificationRepository interface {
	GetByUserId(ctx context.Context, userId uuid.UUID, page int, pageSize int) ([]Notification, int, error)
	GetUnreadCount(ctx context.Context, userId uuid.UUID) (int, error)
	Create(ctx context.Context, notification Notification) (*Notification, error)
	MarkAsRead(ctx context.Context, notificationId uuid.UUID, userId uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userId uuid.UUID) error
	DeleteById(ctx context.Context, notificationId uuid.UUID, userId uuid.UUID) error
}
