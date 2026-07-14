package dto

import "github.com/gofrs/uuid/v5"

type PurchaseRequestDTO struct {
	ShopItemId      uuid.UUID  `json:"shopItemId" validate:"required"`
	Quantity        int64      `json:"quantity" validate:"required,min=1,max=1000"`
	RecipientUserId *uuid.UUID `json:"recipientUserId"`
	Message         *string    `json:"message" validate:"omitempty,max=500"`
}

type PurchaseResponseDTO struct {
	GiftId       *string `json:"giftId"`
	AnimationURL string  `json:"animationUrl"`
}
