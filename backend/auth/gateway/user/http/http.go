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
	serviceresponse "sen1or/letslive/auth/response"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
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

// GetUserStatus looks up a single user's status via the internal batch-status
// endpoint rather than the public profile route, which filters out disabled
// accounts entirely (see get_public_info_by_id.go) and would otherwise make
// this call 404 for exactly the accounts it needs to detect.
func (g *userGateway) GetUserStatus(ctx context.Context, userId string) (string, *serviceresponse.Response[any]) {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	userUUID, err := uuid.FromString(userId)
	if err != nil {
		logger.Errorf(ctx, "invalid user id for status lookup: %s", err)
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(&struct {
		UserIds []uuid.UUID `json:"userIds"`
	}{UserIds: []uuid.UUID{userUUID}}); err != nil {
		logger.Errorf(ctx, "failed to encode statuses request body: %s", err)
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	url := fmt.Sprintf("http://%s/v1/internal/users/statuses", addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, payloadBuf)
	if err != nil {
		logger.Errorf(ctx, "failed to create the request: %s", err)
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}
	req.Header.Set("Content-Type", "application/json")

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Errorf(ctx, "failed to create the request: %s", err)
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(ctx, "failed to call request: %s", err)
		return "", serviceresponse.NewResponseFromTemplate[any](
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
			return "", serviceresponse.NewResponseFromTemplate[any](
				serviceresponse.RES_ERR_INTERNAL_SERVER,
				nil,
				nil,
				nil,
			)
		}

		return "", &resInfo
	}

	var statusRes serviceresponse.Response[dto.GetUsersStatusesResponseDTO]
	if err := json.NewDecoder(resp.Body).Decode(&statusRes); err != nil {
		logger.Errorf(ctx, "failed to decode resp body: %s", err)
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	if statusRes.Data == nil {
		logger.Errorf(ctx, "user service returned no data for status lookup")
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	status, ok := statusRes.Data.Statuses[userId]
	if !ok {
		logger.Errorf(ctx, "user service returned no status entry for user %s", userId)
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	return status, nil
}

func (g *userGateway) UpdateUserStatus(ctx context.Context, userId string, status string) *serviceresponse.Response[any] {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	url := fmt.Sprintf("http://%s/v1/user/%s", addr, userId)
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(&dto.UpdateUserStatusRequestDTO{Status: status}); err != nil {
		logger.Errorf(ctx, "failed to encode user status dto body: %s", err)
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, payloadBuf)
	if err != nil {
		logger.Errorf(ctx, "failed to create the request: %s", err)
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Errorf(ctx, "failed to create the request: %s", err)
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(ctx, "failed to call request: %s", err)
		return serviceresponse.NewResponseFromTemplate[any](
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
			return serviceresponse.NewResponseFromTemplate[any](
				serviceresponse.RES_ERR_INTERNAL_SERVER,
				nil,
				nil,
				nil,
			)
		}

		return &resInfo
	}

	return nil
}
