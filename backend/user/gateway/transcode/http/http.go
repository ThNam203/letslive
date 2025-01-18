package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/lets-live/pkg/discovery"
	"sen1or/lets-live/pkg/logger"
	"sen1or/lets-live/user/dto"
	gateway "sen1or/lets-live/user/gateway"
	transcodegateway "sen1or/lets-live/user/gateway/transcode"
)

type transcodeGateway struct {
	registry discovery.Registry
}

func NewTranscodeGateway(registry discovery.Registry) transcodegateway.TranscodeGateway {
	return &transcodeGateway{
		registry: registry,
	}
}

func (g *transcodeGateway) GetUserVODs(ctx context.Context, userId string) (*dto.TranscodeService_GetUserResponse, *gateway.ErrorResponse) {
	addr, err := g.registry.ServiceAddress(ctx, "transcode")
	if err != nil {
		return nil, &gateway.ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusBadGateway,
		}
	}

	url := fmt.Sprintf("http://%s/v1/vod/%s", addr, userId)

	logger.Infof("transcode gateway url", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
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
				Message:    fmt.Sprintf("failed to decode error response: %s", err),
				StatusCode: http.StatusInternalServerError,
			}
		}

		return nil, &resInfo
	}

	var userVODs dto.TranscodeService_GetUserResponse
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&userVODs); err != nil {
		return nil, &gateway.ErrorResponse{
			Message:    fmt.Sprintf("failed to decode resp body: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return &userVODs, nil
}
