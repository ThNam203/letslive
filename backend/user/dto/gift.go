package dto

type SendGiftRequestDTO struct {
	ShopItemId      string  `json:"shop_item_id" validate:"required,uuid"`
	RecipientUserId string  `json:"recipient_user_id" validate:"required,uuid"`
	Message         *string `json:"message"`
}

type CreateGiftInternalRequestDTO struct {
	SenderId    string  `json:"sender_id" validate:"required,uuid"`
	RecipientId string  `json:"recipient_id" validate:"required,uuid"`
	ShopItemId  string  `json:"shop_item_id" validate:"required,uuid"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	Message     *string `json:"message"`
}
