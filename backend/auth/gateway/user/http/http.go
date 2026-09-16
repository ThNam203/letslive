package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/gateway"
	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/auth/gateway/user/dto"
	serviceresponse "sen1or/letslive/auth/response"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
)

type userGateway struct {
	registry discovery.Registry
}

func NewUserGateway(registry discovery.Registry) usergateway.UserGateway {
	return &userGateway{
		registry: registry,
	}
}

func (g *userGateway) CreateNewUser(ctx context.Context, userRequestDTO dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, error) {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		return nil, domains.ErrInternal
	}

	url := fmt.Sprintf("http://%s/v1/user", addr)
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(&userRequestDTO); err != nil {
		logger.Errorf(ctx, "failed to encode user dto body: %s", err)
		return nil, domains.ErrInternal
	}

	req, err := http.NewRequest(http.MethodPost, url, payloadBuf)
	if err != nil {
		logger.Errorf(ctx, "failed to create the request: %s", err)
		return nil, domains.ErrInternal
	}

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Errorf(ctx, "failed to create the request: %s", err)
		return nil, domains.ErrInternal
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(ctx, "failed to call request: %s", err)
		return nil, domains.ErrInternal
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		var resInfo serviceresponse.Response[any]
		if err := json.NewDecoder(resp.Body).Decode(&resInfo); err != nil {
			logger.Errorf(ctx, "failed to decode error response from user service: %s", err)
			return nil, domains.ErrInternal
		}

		// the downstream code and key must survive: a taken username is
		// reported by the user service, not by auth
		return nil, &domains.DownstreamError{
			StatusCode: resp.StatusCode,
			Code:       resInfo.Code,
			Key:        resInfo.Key,
			Message:    resInfo.Message,
		}
	}

	var createdUser serviceresponse.Response[dto.CreateUserResponseDTO]

	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		logger.Errorf(ctx, "failed to decode resp body: %s", err)
		return nil, domains.ErrInternal
	}

	return createdUser.Data, nil
}
