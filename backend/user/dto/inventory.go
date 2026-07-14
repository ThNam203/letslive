package dto

// AddInventoryInternalRequestDTO is the finance→user internal contract; field names are
// camelCase to match the finance gateway payload (gateway/userservice/http/http.go).
type AddInventoryInternalRequestDTO struct {
	UserId     string `json:"userId" validate:"required,uuid"`
	ShopItemId string `json:"shopItemId" validate:"required,uuid"`
	Quantity   int    `json:"quantity" validate:"required,min=1,max=1000"`
}
