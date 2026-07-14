package purchaseservice

import (
	"context"
	"fmt"

	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/dto"
	"sen1or/letslive/finance/gateway/userservice"
	"sen1or/letslive/finance/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

type PurchaseService struct {
	accountRepo     domains.AccountRepository
	currencyRepo    domains.CurrencyRepository
	transactionRepo domains.TransactionRepository
	shopItemRepo    domains.ShopItemRepository
	userGateway     userservice.UserServiceGateway
}

func NewPurchaseService(
	accountRepo domains.AccountRepository,
	currencyRepo domains.CurrencyRepository,
	transactionRepo domains.TransactionRepository,
	shopItemRepo domains.ShopItemRepository,
	userGateway userservice.UserServiceGateway,
) *PurchaseService {
	return &PurchaseService{
		accountRepo:     accountRepo,
		currencyRepo:    currencyRepo,
		transactionRepo: transactionRepo,
		shopItemRepo:    shopItemRepo,
		userGateway:     userGateway,
	}
}

// Purchase debits the buyer's wallet in the item's currency, then grants the item via
// the user service (inventory, or a gift when RecipientUserId is set). If the grant
// fails after money moved, a compensating refund transaction credits the wallet back.
func (s *PurchaseService) Purchase(ctx context.Context, actorID uuid.UUID, req dto.PurchaseRequestDTO) (*dto.PurchaseResponseDTO, *response.Response[any]) {
	item, serviceErr := s.shopItemRepo.GetById(ctx, req.ShopItemId)
	if serviceErr != nil {
		return nil, serviceErr
	}

	currency, serviceErr := s.currencyRepo.GetByCode(ctx, item.CurrencyCode)
	if serviceErr != nil {
		return nil, serviceErr
	}

	totalCost, costErr := minorUnitCost(item.Price, req.Quantity, currency.Precision)
	if costErr != nil {
		logger.Errorf(ctx, "purchase cost overflow [purchase: %v]", costErr)
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_AMOUNT, nil, nil, nil)
	}

	wallet, serviceErr := s.accountRepo.GetUserWalletByOwnerId(ctx, actorID)
	if serviceErr != nil {
		return nil, serviceErr
	}
	if wallet.Status == domains.AccountStatusFrozen {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_ACCOUNT_FROZEN, nil, nil, nil)
	}

	balances, serviceErr := s.accountRepo.GetBalances(ctx, wallet.Id)
	if serviceErr != nil {
		return nil, serviceErr
	}
	var available int64
	for _, b := range balances {
		if b.CurrencyCode == item.CurrencyCode {
			available = b.Balance
			break
		}
	}
	if available < totalCost {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INSUFFICIENT_BALANCE, nil, nil, nil)
	}

	escrow, serviceErr := s.accountRepo.GetEscrow(ctx)
	if serviceErr != nil {
		return nil, serviceErr
	}

	reference := fmt.Sprintf("purchase-%s", uuid.Must(uuid.NewV4()).String())
	tx, serviceErr := s.transactionRepo.Create(ctx, domains.Transaction{
		Type:      domains.TransactionTypePurchase,
		Reference: &reference,
		Status:    domains.ProcessStatusCreated,
		ActorId:   &actorID,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	entries := []domains.LedgerEntryDraft{
		{AccountId: wallet.Id, CurrencyCode: item.CurrencyCode, Amount: -totalCost},
		{AccountId: escrow.Id, CurrencyCode: item.CurrencyCode, Amount: totalCost},
	}
	if completeErr := s.transactionRepo.CompleteWithEntries(ctx, tx.Id, entries); completeErr != nil {
		return nil, completeErr
	}

	if req.RecipientUserId != nil {
		giftID, err := s.userGateway.CreateGift(ctx, actorID.String(), req.RecipientUserId.String(), item.Id.String(), req.Quantity, req.Message)
		if err != nil {
			logger.Errorf(ctx, "gift grant failed after debit, refunding transaction %s [purchase: %v]", tx.Id, err)
			s.refund(ctx, actorID, tx.Id, wallet.Id, escrow.Id, item.CurrencyCode, totalCost)
			return nil, response.NewResponseFromTemplate[any](response.RES_ERR_USER_SERVICE_ERROR, nil, nil, nil)
		}
		return &dto.PurchaseResponseDTO{GiftId: &giftID, AnimationURL: item.AnimationURL}, nil
	}

	if err := s.userGateway.AddInventory(ctx, actorID.String(), item.Id.String(), req.Quantity); err != nil {
		logger.Errorf(ctx, "inventory grant failed after debit, refunding transaction %s [purchase: %v]", tx.Id, err)
		s.refund(ctx, actorID, tx.Id, wallet.Id, escrow.Id, item.CurrencyCode, totalCost)
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_USER_SERVICE_ERROR, nil, nil, nil)
	}

	return &dto.PurchaseResponseDTO{AnimationURL: item.AnimationURL}, nil
}

// refund is the best-effort compensating action when the item grant fails after the
// wallet was debited. A failure here means stranded funds and must be reconciled
// manually, so it is logged at error level with the original transaction id.
func (s *PurchaseService) refund(ctx context.Context, actorID uuid.UUID, purchaseTxId uuid.UUID, walletId uuid.UUID, escrowId uuid.UUID, currencyCode string, amount int64) {
	reference := fmt.Sprintf("refund-%s", purchaseTxId.String())
	refundTx, serviceErr := s.transactionRepo.Create(ctx, domains.Transaction{
		Type:      domains.TransactionTypeRefund,
		Reference: &reference,
		Status:    domains.ProcessStatusCreated,
		ActorId:   &actorID,
	})
	if serviceErr != nil {
		logger.Errorf(ctx, "CRITICAL: refund creation failed for purchase %s, funds stranded in escrow [refund]", purchaseTxId)
		return
	}

	entries := []domains.LedgerEntryDraft{
		{AccountId: walletId, CurrencyCode: currencyCode, Amount: amount},
		{AccountId: escrowId, CurrencyCode: currencyCode, Amount: -amount},
	}
	if completeErr := s.transactionRepo.CompleteWithEntries(ctx, refundTx.Id, entries); completeErr != nil {
		logger.Errorf(ctx, "CRITICAL: refund completion failed for purchase %s (refund tx %s), funds stranded in escrow [refund]", purchaseTxId, refundTx.Id)
	}
}

// minorUnitCost converts a whole-unit item price into minor units (price * quantity *
// 10^precision) with overflow checks on every multiplication.
func minorUnitCost(price int64, quantity int64, precision int) (int64, error) {
	if price <= 0 || quantity <= 0 {
		return 0, fmt.Errorf("price and quantity must be positive")
	}

	factor := int64(1)
	for i := 0; i < precision; i++ {
		factor *= 10
	}

	subtotal := price * quantity
	if subtotal/quantity != price {
		return 0, fmt.Errorf("price*quantity overflows")
	}
	total := subtotal * factor
	if total/factor != subtotal {
		return 0, fmt.Errorf("cost in minor units overflows")
	}
	return total, nil
}
