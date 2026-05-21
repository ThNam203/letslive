package deposit

import (
	"sen1or/letslive/finance/handlers/basehandler"
	"sen1or/letslive/finance/services/deposit"
)

type DepositHandler struct {
	basehandler.BaseHandler
	depositService *deposit.DepositService
}

func NewDepositHandler(depositService *deposit.DepositService) *DepositHandler {
	return &DepositHandler{
		depositService: depositService,
	}
}
