package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/lets-live/pkg/discovery"
	"sen1or/lets-live/user/dto"
)

type UserGateway struct {
	registry discovery.Registry
}

func NewUserGateway(registry discovery.Registry) *UserGateway {
	return &UserGateway{
		registry: registry,
	}
}

func (g *UserGateway) CreateNewUser(ctx context.Context, userRequestDTO dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, error) {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://%s/v1/user", addr)
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(&userRequestDTO); err != nil {
		return nil, fmt.Errorf("failed to encode user dto body: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, payloadBuf)
	if err != nil {
		return nil, err
	}

	resq, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call request: %s", err)
	}
	defer resq.Body.Close()

	if resq.StatusCode/100 != 2 {
		return nil, fmt.Errorf("failed calling request for creating user: error code %d", resq.StatusCode)
	}

	var createdUser dto.CreateUserResponseDTO
	defer resq.Body.Close()

	if err := json.NewDecoder(resq.Body).Decode(&createdUser); err != nil {
		return nil, fmt.Errorf("failed to decode resp body: %s", err)
	}

	return &createdUser, nil
}
