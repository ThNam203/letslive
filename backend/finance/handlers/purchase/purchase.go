package purchasehandler

import (
	"sen1or/letslive/finance/handlers/basehandler"
	purchaseservice "sen1or/letslive/finance/services/purchase"
)

type PurchaseHandler struct {
	basehandler.BaseHandler
	purchaseService *purchaseservice.PurchaseService
}

func NewPurchaseHandler(purchaseService *purchaseservice.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{purchaseService: purchaseService}
}
