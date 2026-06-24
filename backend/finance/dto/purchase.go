package dto

import "github.com/google/uuid"

type PurchaseRequestDTO struct {
	ShopItemId      uuid.UUID  `json:"shopItemId" validate:"required"`
	Quantity        int64      `json:"quantity" validate:"required,min=1"`
	RecipientUserId *uuid.UUID `json:"recipientUserId"`
	Message         *string    `json:"message"`
}

type PurchaseResponseDTO struct {
	GiftId       *string `json:"giftId"`
	AnimationURL string  `json:"animationUrl"`
}
