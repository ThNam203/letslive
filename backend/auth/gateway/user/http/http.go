package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/letslive/auth/gateway"
	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/auth/pkg/discovery"
)

type userGateway struct {
	registry discovery.Registry
}

func NewUserGateway(registry discovery.Registry) usergateway.UserGateway {
	return &userGateway{
		registry: registry,
	}
}

func (g *userGateway) CreateNewUser(ctx context.Context, userRequestDTO usergateway.CreateUserRequestDTO) (*usergateway.CreateUserResponseDTO, *gateway.ErrorResponse) {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		return nil, &gateway.ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusBadGateway,
		}
	}

	url := fmt.Sprintf("http://%s/v1/user", addr)
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(&userRequestDTO); err != nil {
		return nil, &gateway.ErrorResponse{
			Message:    fmt.Sprintf("failed to encode user dto body: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	req, err := http.NewRequest(http.MethodPost, url, payloadBuf)
	if err != nil {
		return nil, &gateway.ErrorResponse{
			Message:    fmt.Sprintf("failed to create the request: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &gateway.ErrorResponse{
			Message:    fmt.Sprintf("failed to call request: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		resInfo := gateway.ErrorResponse{}
		if err := json.NewDecoder(resp.Body).Decode(&resInfo); err != nil {
			return nil, &gateway.ErrorResponse{
				Message:    fmt.Sprintf("failed to decode error response from user service: %s", err),
				StatusCode: http.StatusInternalServerError,
			}
		}

		return nil, &resInfo
	}

	var createdUser usergateway.CreateUserResponseDTO
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		return nil, &gateway.ErrorResponse{
			Message:    fmt.Sprintf("failed to decode resp body: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return &createdUser, nil
}

func (g *userGateway) UpdateUserVerified(ctx context.Context, userId string) *gateway.ErrorResponse {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		return &gateway.ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusBadGateway,
		}
	}

	url := fmt.Sprintf("http://%s/v1/user/%s/set-verified", addr, userId)

	req, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return &gateway.ErrorResponse{
			Message:    fmt.Sprintf("failed to create the request: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &gateway.ErrorResponse{
			Message:    fmt.Sprintf("failed to call request: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		resInfo := gateway.ErrorResponse{}
		if err := json.NewDecoder(resp.Body).Decode(&resInfo); err != nil {
			return &gateway.ErrorResponse{
				Message:    fmt.Sprintf("failed to decode error response from user service: %s", err),
				StatusCode: http.StatusInternalServerError,
			}
		}

		return &resInfo
	}

	return nil
}
