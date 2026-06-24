package domains

import (
	"context"
	"time"

	"github.com/google/uuid"
	"sen1or/letslive/finance/response"
)

type ShopItem struct {
	Id           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  *string   `json:"description" db:"description"`
	ImageURL     string    `json:"imageUrl" db:"image_url"`
	AnimationURL string    `json:"animationUrl" db:"animation_url"`
	Price        int64     `json:"price" db:"price"`
	IsActive     bool      `json:"isActive" db:"is_active"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

type ShopItemRepository interface {
	List(ctx context.Context) ([]ShopItem, *response.Response[any])
	GetById(ctx context.Context, id uuid.UUID) (*ShopItem, *response.Response[any])
}
