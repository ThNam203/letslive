package finance

import "context"

type ShopItem struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	AnimationURL string `json:"animationUrl"`
}

type FinanceGateway interface {
	GetShopItem(ctx context.Context, shopItemID string) (*ShopItem, error)
}
