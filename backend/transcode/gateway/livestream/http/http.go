package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/lets-live/pkg/discovery"
	"sen1or/lets-live/transcode/dto"
)

type LivestreamGateway struct {
	registry discovery.Registry
}

type ErrorResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

func NewLivestreamGateway(registry discovery.Registry) *LivestreamGateway {
	return &LivestreamGateway{
		registry: registry,
	}
}

func (g *LivestreamGateway) Create(ctx context.Context, data dto.CreateLivestreamRequestDTO) (*dto.LivestreamResponseDTO, *ErrorResponse) {
	addr, err := g.registry.ServiceAddress(ctx, "livestream")
	if err != nil {
		return nil, &ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusBadGateway,
		}
	}

	url := fmt.Sprintf("http://%s/v1/livestream", addr)
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(data); err != nil {
		return nil, &ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}

	}
	req, err := http.NewRequest(http.MethodPost, url, payloadBuf)
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

	var livestreamResponse dto.LivestreamResponseDTO

	if err := json.NewDecoder(resp.Body).Decode(&livestreamResponse); err != nil {
		return nil, &ErrorResponse{
			Message:    fmt.Sprintf("failed to decode resp body: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return &livestreamResponse, nil
}

func (g *LivestreamGateway) Update(ctx context.Context, updateDTO dto.UpdateLivestreamRequestDTO) *ErrorResponse {
	addr, err := g.registry.ServiceAddress(ctx, "livestream")
	if err != nil {
		return &ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusBadGateway,
		}
	}

	url := fmt.Sprintf("http://%s/v1/livestream/%s", addr, updateDTO.Id)
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(updateDTO); err != nil {
		return &ErrorResponse{
			Message:    fmt.Sprintf("failed to encode livestream body: %s", err),
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
			Message:    fmt.Sprintf("failed to update livestream from livestream service: %s", err),
			StatusCode: resp.StatusCode,
		}
	}

	return nil
}
