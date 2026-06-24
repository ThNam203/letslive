package shopitemhandler

import (
	"sen1or/letslive/finance/handlers/basehandler"
	shopitemservice "sen1or/letslive/finance/services/shop_item"
)

type ShopItemHandler struct {
	basehandler.BaseHandler
	shopItemService *shopitemservice.ShopItemService
}

func NewShopItemHandler(shopItemService *shopitemservice.ShopItemService) *ShopItemHandler {
	return &ShopItemHandler{shopItemService: shopItemService}
}
