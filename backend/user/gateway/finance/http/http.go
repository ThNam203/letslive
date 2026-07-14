package financehttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/gateway"
	financegateway "sen1or/letslive/user/gateway/finance"
)

type financeHTTPGateway struct {
	registry discovery.Registry
}

func NewFinanceGateway(registry discovery.Registry) financegateway.FinanceGateway {
	return &financeHTTPGateway{
		registry: registry,
	}
}

type getShopItemResponse struct {
	Data *financegateway.ShopItem `json:"data"`
}

func (g *financeHTTPGateway) GetShopItem(ctx context.Context, shopItemID string) (*financegateway.ShopItem, error) {
	addr, err := g.registry.ServiceAddress(ctx, "finance")
	if err != nil {
		logger.Errorf(ctx, "failed to get finance service address: %v", err)
		return nil, fmt.Errorf("finance service unavailable")
	}

	url := fmt.Sprintf("http://%s/v1/shop/items/%s", addr, shopItemID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Warnf(ctx, "failed to set request id header: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(ctx, "failed to call finance service GetShopItem: %v", err)
		return nil, fmt.Errorf("failed to call finance service")
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("finance service returned status %d on GetShopItem", resp.StatusCode)
	}

	var result getShopItemResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Errorf(ctx, "failed to decode GetShopItem response: %v", err)
		return nil, fmt.Errorf("failed to decode finance service response")
	}
	if result.Data == nil {
		return nil, fmt.Errorf("finance service returned empty shop item")
	}

	return result.Data, nil
}
