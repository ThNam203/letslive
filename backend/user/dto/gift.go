package dto

type SendGiftRequestDTO struct {
	ShopItemId      string  `json:"shopItemId" validate:"required,uuid"`
	RecipientUserId string  `json:"recipientUserId" validate:"required,uuid"`
	Message         *string `json:"message" validate:"omitempty,max=500"`
}

// CreateGiftInternalRequestDTO is the finance→user internal contract; field names are
// camelCase to match the finance gateway payload (gateway/userservice/http/http.go).
type CreateGiftInternalRequestDTO struct {
	SenderId    string  `json:"senderId" validate:"required,uuid"`
	RecipientId string  `json:"recipientId" validate:"required,uuid"`
	ShopItemId  string  `json:"shopItemId" validate:"required,uuid"`
	Quantity    int     `json:"quantity" validate:"required,min=1,max=1000"`
	Message     *string `json:"message" validate:"omitempty,max=500"`
}
