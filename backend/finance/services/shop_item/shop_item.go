package shopitemservice

import (
	"context"

	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/response"

	"github.com/google/uuid"
)

type ShopItemService struct {
	shopItemRepo domains.ShopItemRepository
}

func NewShopItemService(shopItemRepo domains.ShopItemRepository) *ShopItemService {
	return &ShopItemService{shopItemRepo: shopItemRepo}
}

func (s *ShopItemService) List(ctx context.Context) ([]domains.ShopItem, *response.Response[any]) {
	return s.shopItemRepo.List(ctx)
}

func (s *ShopItemService) GetById(ctx context.Context, id uuid.UUID) (*domains.ShopItem, *response.Response[any]) {
	return s.shopItemRepo.GetById(ctx, id)
}
