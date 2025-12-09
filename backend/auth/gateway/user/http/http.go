package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/letslive/auth/gateway"
	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/auth/gateway/user/dto"
	"sen1or/letslive/auth/pkg/discovery"
	"sen1or/letslive/auth/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"
)

type userGateway struct {
	registry discovery.Registry
}

func NewUserGateway(registry discovery.Registry) usergateway.UserGateway {
	return &userGateway{
		registry: registry,
	}
}

func (g *userGateway) CreateNewUser(ctx context.Context, userRequestDTO dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, *serviceresponse.Response[any]) {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	url := fmt.Sprintf("http://%s/v1/user", addr)
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(&userRequestDTO); err != nil {
		logger.Errorf(ctx, "failed to encode user dto body: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	req, err := http.NewRequest(http.MethodPost, url, payloadBuf)
	if err != nil {
		logger.Errorf(ctx, "failed to create the request: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Errorf(ctx, "failed to create the request: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(ctx, "failed to call request: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		resInfo := serviceresponse.Response[any]{}
		if err := json.NewDecoder(resp.Body).Decode(&resInfo); err != nil {
			logger.Errorf(ctx, "failed to decode error response from user service: %s", err)
			return nil, serviceresponse.NewResponseFromTemplate[any](
				serviceresponse.RES_ERR_INTERNAL_SERVER,
				nil,
				nil,
				nil,
			)
		}

		return nil, &resInfo
	}

	var createdUser serviceresponse.Response[dto.CreateUserResponseDTO]
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		logger.Errorf(ctx, "failed to decode resp body: %s", err)
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	return createdUser.Data, nil
}
