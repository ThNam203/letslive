package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/lets-live/pkg/discovery"
	dto "sen1or/lets-live/transcode/gateway/user"
)

type UserGateway struct {
	registry discovery.Registry
}

type ErrorResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

func NewUserGateway(registry discovery.Registry) *UserGateway {
	return &UserGateway{
		registry: registry,
	}
}

func (g *UserGateway) GetUserInformation(ctx context.Context, streamAPIKey string) (*dto.GetUserResponseDTO, *ErrorResponse) {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		return nil, &ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusBadGateway,
		}
	}

	url := fmt.Sprintf("http://%s/v1/verify-stream-key?streamAPIKey=%s", addr, streamAPIKey)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, &ErrorResponse{
			Message:    fmt.Sprintf("failed to create request: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &ErrorResponse{
			Message:    fmt.Sprintf("failed to call request: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		resInfo := ErrorResponse{}
		if err := json.NewDecoder(resp.Body).Decode(&resInfo); err != nil {
			return nil, &ErrorResponse{
				Message:    fmt.Sprintf("failed to decode error response from user service: %s", err),
				StatusCode: http.StatusInternalServerError,
			}
		}

		return nil, &resInfo
	}

	var userInfo dto.GetUserResponseDTO

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, &ErrorResponse{
			Message:    fmt.Sprintf("failed to decode resp body: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return &userInfo, nil
}

func (g *UserGateway) UpdateUserLiveStatus(ctx context.Context, dto dto.UpdateUserLiveStatusDTO) *ErrorResponse {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		return &ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusBadGateway,
		}
	}

	url := fmt.Sprintf("http://%s/v1/user/%s", addr, dto.Id)
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(dto); err != nil {
		return &ErrorResponse{
			Message:    fmt.Sprintf("failed to encode user body: %s", err),
			StatusCode: 500,
		}
	}

	req, err := http.NewRequest(http.MethodPut, url, payloadBuf)
	if err != nil {
		return &ErrorResponse{
			Message:    fmt.Sprintf("failed to create request: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &ErrorResponse{
			Message:    fmt.Sprintf("failed to call request: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return &ErrorResponse{
			Message:    fmt.Sprintf("failed to update user from user service: %s", err),
			StatusCode: resp.StatusCode,
		}
	}

	return nil
}
