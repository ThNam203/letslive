package domains

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

type UserInventory struct {
	Id         uuid.UUID `json:"id" db:"id"`
	UserId     uuid.UUID `json:"userId" db:"user_id"`
	ShopItemId uuid.UUID `json:"shopItemId" db:"shop_item_id"`
	Quantity   int       `json:"quantity" db:"quantity"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

type InventoryRepository interface {
	Upsert(ctx context.Context, userID, shopItemID uuid.UUID, quantityToAdd int) (*UserInventory, error)
	Deduct(ctx context.Context, userID, shopItemID uuid.UUID) (*UserInventory, error)
	GetByUserId(ctx context.Context, userID uuid.UUID, page, limit int) ([]UserInventory, int, error)
}
