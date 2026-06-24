package userservice

import "context"

type UserServiceGateway interface {
	AddInventory(ctx context.Context, userID, shopItemID string, quantity int64) error
	CreateGift(ctx context.Context, senderID, recipientID, shopItemID string, quantity int64, message *string) (giftID string, err error)
}
