package domains

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"sen1or/letslive/user/response"
)

type Gift struct {
	Id              uuid.UUID `json:"id" db:"id"`
	SenderUserId    uuid.UUID `json:"senderUserId" db:"sender_user_id"`
	RecipientUserId uuid.UUID `json:"recipientUserId" db:"recipient_user_id"`
	ShopItemId      uuid.UUID `json:"shopItemId" db:"shop_item_id"`
	Quantity        int       `json:"quantity" db:"quantity"`
	Message         *string   `json:"message" db:"message"`
	SentAt          time.Time `json:"sentAt" db:"sent_at"`
}

type GiftRepository interface {
	Create(ctx context.Context, gift Gift) (*Gift, *response.Response[any])
	ListByRecipient(ctx context.Context, recipientID uuid.UUID, page, limit int) ([]Gift, int, *response.Response[any])
	ListBySender(ctx context.Context, senderID uuid.UUID, page, limit int) ([]Gift, int, *response.Response[any])
}
