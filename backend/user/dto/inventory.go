package dto

type AddInventoryInternalRequestDTO struct {
	UserId     string `json:"user_id" validate:"required,uuid"`
	ShopItemId string `json:"shop_item_id" validate:"required,uuid"`
	Quantity   int    `json:"quantity" validate:"required,min=1"`
}
