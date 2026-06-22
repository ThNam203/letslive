package wallet

import (
	"sen1or/letslive/finance/handlers/basehandler"
	"sen1or/letslive/finance/services/wallet"
)

type WalletHandler struct {
	basehandler.BaseHandler
	walletService *wallet.WalletService
}

func NewWalletHandler(walletService *wallet.WalletService) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}
