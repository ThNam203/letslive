package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type ShopItemResponseDTO struct {
	Id           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	ImageURL     string    `json:"imageUrl"`
	AnimationURL string    `json:"animationUrl"`
	Price        int64     `json:"price"` // whole units of CurrencyCode
	CurrencyCode string    `json:"currencyCode"`
	CreatedAt    time.Time `json:"createdAt"`
}
