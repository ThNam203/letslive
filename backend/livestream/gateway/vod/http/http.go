package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/letslive/livestream/gateway"
	vodgateway "sen1or/letslive/livestream/gateway/vod"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

type vodHTTPGateway struct {
	registry discovery.Registry
}

func NewVODGateway(registry discovery.Registry) vodgateway.VODGateway {
	return &vodHTTPGateway{
		registry: registry,
	}
}

type vodCreateResponse struct {
	Success bool `json:"success"`
	Data    *struct {
		Id uuid.UUID `json:"id"`
	} `json:"data,omitempty"`
}

func (g *vodHTTPGateway) CreateVOD(ctx context.Context, req vodgateway.CreateVODRequest) (*uuid.UUID, error) {
	addr, err := g.registry.ServiceAddress(ctx, "vod")
	if err != nil {
		logger.Errorf(ctx, "failed to get vod service address: %v", err)
		return nil, fmt.Errorf("vod service unavailable")
	}

	url := fmt.Sprintf("http://%s/v1/internal/vods", addr)
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(req); err != nil {
		return nil, fmt.Errorf("failed to encode vod create request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, payloadBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	if err := gateway.SetRequestIDHeader(ctx, httpReq); err != nil {
		logger.Warnf(ctx, "failed to set request id header: %v", err)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call vod service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("vod service returned status %d", resp.StatusCode)
	}

	var result vodCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode vod service response: %w", err)
	}

	if result.Data != nil {
		return &result.Data.Id, nil
	}

	return nil, nil
}
