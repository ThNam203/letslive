package purchaseservice

import (
	"context"
	"fmt"

	"sen1or/letslive/finance/domains"
	"sen1or/letslive/finance/dto"
	"sen1or/letslive/finance/gateway/userservice"
	"sen1or/letslive/finance/response"

	"github.com/gofrs/uuid/v5"
	googleuuid "github.com/google/uuid"
)

type PurchaseService struct {
	accountRepo     domains.AccountRepository
	transactionRepo domains.TransactionRepository
	shopItemRepo    domains.ShopItemRepository
	userGateway     userservice.UserServiceGateway
}

func NewPurchaseService(
	accountRepo domains.AccountRepository,
	transactionRepo domains.TransactionRepository,
	shopItemRepo domains.ShopItemRepository,
	userGateway userservice.UserServiceGateway,
) *PurchaseService {
	return &PurchaseService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		shopItemRepo:    shopItemRepo,
		userGateway:     userGateway,
	}
}

func (s *PurchaseService) Purchase(ctx context.Context, actorID uuid.UUID, req dto.PurchaseRequestDTO) (*dto.PurchaseResponseDTO, *response.Response[any]) {
	shopItemID, err := googleuuid.Parse(req.ShopItemId.String())
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_SHOP_ITEM_NOT_FOUND, nil, nil, nil)
	}

	item, serviceErr := s.shopItemRepo.GetById(ctx, shopItemID)
	if serviceErr != nil {
		return nil, serviceErr
	}

	totalCost := item.Price * req.Quantity

	wallet, serviceErr := s.accountRepo.GetUserWalletByOwnerId(ctx, actorID)
	if serviceErr != nil {
		return nil, serviceErr
	}

	balances, serviceErr := s.accountRepo.GetBalances(ctx, wallet.Id)
	if serviceErr != nil {
		return nil, serviceErr
	}
	if len(balances) == 0 {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INSUFFICIENT_BALANCE, nil, nil, nil)
	}

	var userBalance int64
	currencyCode := balances[0].CurrencyCode
	for _, b := range balances {
		userBalance += b.Balance
	}
	if userBalance < totalCost {
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
		{AccountId: wallet.Id, CurrencyCode: currencyCode, Amount: -totalCost},
		{AccountId: escrow.Id, CurrencyCode: currencyCode, Amount: totalCost},
	}
	if completeErr := s.transactionRepo.CompleteWithEntries(ctx, tx.Id, entries); completeErr != nil {
		return nil, completeErr
	}

	if req.RecipientUserId != nil {
		giftID, err := s.userGateway.CreateGift(ctx, actorID.String(), req.RecipientUserId.String(), item.Id.String(), req.Quantity, req.Message)
		if err != nil {
			return nil, response.NewResponseFromTemplate[any](response.RES_ERR_USER_SERVICE_ERROR, nil, nil, nil)
		}
		return &dto.PurchaseResponseDTO{GiftId: &giftID, AnimationURL: item.AnimationURL}, nil
	}

	if err := s.userGateway.AddInventory(ctx, actorID.String(), item.Id.String(), req.Quantity); err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_USER_SERVICE_ERROR, nil, nil, nil)
	}

	return &dto.PurchaseResponseDTO{AnimationURL: item.AnimationURL}, nil
}
