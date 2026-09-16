package shopitemservice

import (
	"context"

	"sen1or/letslive/finance/domains"

	"github.com/gofrs/uuid/v5"
)

type ShopItemService struct {
	shopItemRepo domains.ShopItemRepository
}

func NewShopItemService(shopItemRepo domains.ShopItemRepository) *ShopItemService {
	return &ShopItemService{shopItemRepo: shopItemRepo}
}

func (s *ShopItemService) List(ctx context.Context) ([]domains.ShopItem, error) {
	return s.shopItemRepo.List(ctx)
}

func (s *ShopItemService) GetById(ctx context.Context, id uuid.UUID) (*domains.ShopItem, error) {
	return s.shopItemRepo.GetById(ctx, id)
}
