package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/letslive/livestream/gateway"
	usergateway "sen1or/letslive/livestream/gateway/user"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

type userHTTPGateway struct {
	registry discovery.Registry
}

func NewUserGateway(registry discovery.Registry) usergateway.UserGateway {
	return &userHTTPGateway{
		registry: registry,
	}
}

type userServiceResponse struct {
	Success bool                        `json:"success"`
	Data    *usergateway.UserPublicInfo `json:"data,omitempty"`
}

func (g *userHTTPGateway) GetUserPublicInfo(ctx context.Context, userId uuid.UUID) (*usergateway.UserPublicInfo, error) {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		logger.Errorf(ctx, "failed to get user service address: %v", err)
		return nil, fmt.Errorf("user service unavailable")
	}

	url := fmt.Sprintf("http://%s/v1/user/%s", addr, userId.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Errorf(ctx, "failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request")
	}

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Warnf(ctx, "failed to set request id header: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(ctx, "failed to call user service: %v", err)
		return nil, fmt.Errorf("failed to call user service")
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	var result userServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Errorf(ctx, "failed to decode user service response: %v", err)
		return nil, fmt.Errorf("failed to decode user service response")
	}

	return result.Data, nil
}

type getUsersStatusesRequest struct {
	UserIds []uuid.UUID `json:"userIds"`
}

type getUsersStatusesResponse struct {
	Success bool                  `json:"success"`
	Data    *getUsersStatusesData `json:"data,omitempty"`
}

type getUsersStatusesData struct {
	Statuses map[string]string `json:"statuses"`
}

func (g *userHTTPGateway) GetUsersStatuses(ctx context.Context, userIds []uuid.UUID) (map[string]string, error) {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		logger.Errorf(ctx, "failed to get user service address: %v", err)
		return nil, fmt.Errorf("user service unavailable")
	}

	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(&getUsersStatusesRequest{UserIds: userIds}); err != nil {
		logger.Errorf(ctx, "failed to encode statuses request: %v", err)
		return nil, fmt.Errorf("failed to encode request")
	}

	url := fmt.Sprintf("http://%s/v1/internal/users/statuses", addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, payloadBuf)
	if err != nil {
		logger.Errorf(ctx, "failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Warnf(ctx, "failed to set request id header: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(ctx, "failed to call user service: %v", err)
		return nil, fmt.Errorf("failed to call user service")
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	var result getUsersStatusesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Errorf(ctx, "failed to decode user service response: %v", err)
		return nil, fmt.Errorf("failed to decode user service response")
	}
	if result.Data == nil {
		return nil, fmt.Errorf("user service returned no data")
	}

	return result.Data.Statuses, nil
}
