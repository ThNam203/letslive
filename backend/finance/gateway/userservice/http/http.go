package userservicehttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/letslive/finance/gateway"
	"sen1or/letslive/finance/gateway/userservice"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
)

type userServiceHTTPGateway struct {
	registry discovery.Registry
}

func NewUserServiceGateway(registry discovery.Registry) userservice.UserServiceGateway {
	return &userServiceHTTPGateway{
		registry: registry,
	}
}

type addInventoryRequest struct {
	UserId     string `json:"userId"`
	ShopItemId string `json:"shopItemId"`
	Quantity   int64  `json:"quantity"`
}

type createGiftRequest struct {
	SenderId    string  `json:"senderId"`
	RecipientId string  `json:"recipientId"`
	ShopItemId  string  `json:"shopItemId"`
	Quantity    int64   `json:"quantity"`
	Message     *string `json:"message"`
}

type createGiftResponse struct {
	Data struct {
		GiftId string `json:"giftId"`
	} `json:"data"`
}

func (g *userServiceHTTPGateway) AddInventory(ctx context.Context, userID, shopItemID string, quantity int64) error {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		logger.Errorf(ctx, "failed to get user service address: %v", err)
		return fmt.Errorf("user service unavailable")
	}

	body, err := json.Marshal(addInventoryRequest{
		UserId:     userID,
		ShopItemId: shopItemID,
		Quantity:   quantity,
	})
	if err != nil {
		logger.Errorf(ctx, "failed to marshal AddInventory request: %v", err)
		return fmt.Errorf("failed to marshal request")
	}

	url := fmt.Sprintf("http://%s/v1/internal/inventory/add", addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		logger.Errorf(ctx, "failed to create AddInventory request: %v", err)
		return fmt.Errorf("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Warnf(ctx, "failed to set request id header: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(ctx, "failed to call user service AddInventory: %v", err)
		return fmt.Errorf("failed to call user service")
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("user service returned status %d on AddInventory", resp.StatusCode)
	}

	return nil
}

func (g *userServiceHTTPGateway) CreateGift(ctx context.Context, senderID, recipientID, shopItemID string, quantity int64, message *string) (string, error) {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		logger.Errorf(ctx, "failed to get user service address: %v", err)
		return "", fmt.Errorf("user service unavailable")
	}

	body, err := json.Marshal(createGiftRequest{
		SenderId:    senderID,
		RecipientId: recipientID,
		ShopItemId:  shopItemID,
		Quantity:    quantity,
		Message:     message,
	})
	if err != nil {
		logger.Errorf(ctx, "failed to marshal CreateGift request: %v", err)
		return "", fmt.Errorf("failed to marshal request")
	}

	url := fmt.Sprintf("http://%s/v1/internal/gifts/create", addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		logger.Errorf(ctx, "failed to create CreateGift request: %v", err)
		return "", fmt.Errorf("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Warnf(ctx, "failed to set request id header: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(ctx, "failed to call user service CreateGift: %v", err)
		return "", fmt.Errorf("failed to call user service")
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("user service returned status %d on CreateGift", resp.StatusCode)
	}

	var result createGiftResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Errorf(ctx, "failed to decode CreateGift response: %v", err)
		return "", fmt.Errorf("failed to decode user service response")
	}

	return result.Data.GiftId, nil
}
