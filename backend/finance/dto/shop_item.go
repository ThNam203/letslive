package dto

import (
	"time"

	"github.com/google/uuid"
)

type ShopItemResponseDTO struct {
	Id           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	ImageURL     string    `json:"imageUrl"`
	AnimationURL string    `json:"animationUrl"`
	Price        int64     `json:"price"`
	CreatedAt    time.Time `json:"createdAt"`
}
