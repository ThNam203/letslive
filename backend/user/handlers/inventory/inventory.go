package inventoryhandler

import (
	"sen1or/letslive/user/handlers/basehandler"
	"sen1or/letslive/user/services"
)

type InventoryHandler struct {
	basehandler.BaseHandler
	inventoryService *services.InventoryService
}

func NewInventoryHandler(inventoryService *services.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}
