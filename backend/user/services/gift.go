package services

import (
	"context"
	"fmt"

	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

type GiftService struct {
	giftRepo            domains.GiftRepository
	inventoryRepo       domains.InventoryRepository
	notificationService *NotificationService
}

func NewGiftService(giftRepo domains.GiftRepository, inventoryRepo domains.InventoryRepository, notificationService *NotificationService) *GiftService {
	return &GiftService{
		giftRepo:            giftRepo,
		inventoryRepo:       inventoryRepo,
		notificationService: notificationService,
	}
}

// SendFromInventory deducts 1 item from sender's inventory, creates a gift record.
func (s *GiftService) SendFromInventory(ctx context.Context, senderID uuid.UUID, req dto.SendGiftRequestDTO) (*domains.Gift, *response.Response[any]) {
	recipientID, err := uuid.FromString(req.RecipientUserId)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil)
	}
	shopItemID, err := uuid.FromString(req.ShopItemId)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil)
	}

	if _, serviceErr := s.inventoryRepo.Deduct(ctx, senderID, shopItemID); serviceErr != nil {
		return nil, serviceErr
	}

	gift, serviceErr := s.giftRepo.Create(ctx, domains.Gift{
		SenderUserId:    senderID,
		RecipientUserId: recipientID,
		ShopItemId:      shopItemID,
		Quantity:        1,
		Message:         req.Message,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	s.notifyRecipient(ctx, gift)
	return gift, nil
}

// CreateFromPurchase used by internal finance→user quick-send call.
func (s *GiftService) CreateFromPurchase(ctx context.Context, senderID, recipientID, shopItemID uuid.UUID, quantity int, message *string) (*domains.Gift, *response.Response[any]) {
	gift, serviceErr := s.giftRepo.Create(ctx, domains.Gift{
		SenderUserId:    senderID,
		RecipientUserId: recipientID,
		ShopItemId:      shopItemID,
		Quantity:        quantity,
		Message:         message,
	})
	if serviceErr != nil {
		return nil, serviceErr
	}

	s.notifyRecipient(ctx, gift)
	return gift, nil
}

func (s *GiftService) GetReceived(ctx context.Context, recipientID uuid.UUID, page, limit int) ([]domains.Gift, int, *response.Response[any]) {
	return s.giftRepo.ListByRecipient(ctx, recipientID, page, limit)
}

func (s *GiftService) GetSent(ctx context.Context, senderID uuid.UUID, page, limit int) ([]domains.Gift, int, *response.Response[any]) {
	return s.giftRepo.ListBySender(ctx, senderID, page, limit)
}

func (s *GiftService) notifyRecipient(ctx context.Context, gift *domains.Gift) {
	actionURL := "/user/me/gifts/received"
	refIDStr := gift.Id.String()
	s.notificationService.CreateNotification(ctx, dto.CreateNotificationRequestDTO{
		UserId:      gift.RecipientUserId.String(),
		Type:        domains.NotificationTypeGiftReceived,
		Title:       "You received a gift!",
		Message:     fmt.Sprintf("Someone sent you a gift"),
		ActionUrl:   &actionURL,
		ReferenceId: &refIDStr,
	})
}
