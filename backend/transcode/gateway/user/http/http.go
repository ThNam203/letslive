package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/letslive/transcode/gateway"
	"sen1or/letslive/transcode/pkg/discovery"
	"sen1or/letslive/transcode/pkg/logger"
	"sen1or/letslive/transcode/response"
)

type HTTPUserGateway struct {
	registry discovery.Registry
}

func NewUserGateway(registry discovery.Registry) *HTTPUserGateway {
	return &HTTPUserGateway{
		registry: registry,
	}
}

func (g *HTTPUserGateway) GetUserInformation(ctx context.Context, streamAPIKey string) (res *response.Response[GetUserResponseDTO], callErr *response.Response[any]) {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		logger.Debugf(ctx, "get service address from gateway failed")
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}

	url := fmt.Sprintf("http://%s/v1/verify-stream-key?streamAPIKey=%s", addr, streamAPIKey)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Debugf(ctx, "failed to create request: %s", err)
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Debugf(ctx, "failed to set request id header: %s", err)
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Debugf(ctx, "failed to call request: %s", err)
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		resInfo := response.Response[any]{}
		if err := json.NewDecoder(resp.Body).Decode(&resInfo); err != nil {
			logger.Debugf(ctx, "failed to decode error response from user service: %s", err)
			return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
		}

		return nil, &resInfo
	}

	var userInfo response.Response[GetUserResponseDTO]

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		logger.Debugf(ctx, "failed to decode resp body: %s", err)
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}

	return &userInfo, nil
}
