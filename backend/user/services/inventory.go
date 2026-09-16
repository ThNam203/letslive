package services

import (
	"context"

	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
)

type InventoryService struct {
	inventoryRepo domains.InventoryRepository
}

func NewInventoryService(inventoryRepo domains.InventoryRepository) *InventoryService {
	return &InventoryService{inventoryRepo: inventoryRepo}
}

func (s *InventoryService) GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domains.UserInventory, int, error) {
	return s.inventoryRepo.GetByUserId(ctx, userID, page, limit)
}

func (s *InventoryService) AddItems(ctx context.Context, userID, shopItemID uuid.UUID, quantity int) (*domains.UserInventory, error) {
	return s.inventoryRepo.Upsert(ctx, userID, shopItemID, quantity)
}
