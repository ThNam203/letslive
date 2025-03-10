package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sen1or/lets-live/pkg/discovery"
	gateway "sen1or/lets-live/user/gateway"
	livestreamgateway "sen1or/lets-live/user/gateway/livestream"
	"strconv"
	"strings"
)

type livestreamGateway struct {
	registry discovery.Registry
}

func NewLivestreamGateway(registry discovery.Registry) livestreamgateway.LivestreamGateway {
	return &livestreamGateway{
		registry: registry,
	}
}

func (g *livestreamGateway) GetUserLivestreams(ctx context.Context, userId string) ([]livestreamgateway.GetLivestreamResponseDTO, *gateway.ErrorResponse) {
	addr, err := g.registry.ServiceAddress(ctx, "livestream")
	if err != nil {
		return nil, &gateway.ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusBadGateway,
		}
	}

	url := fmt.Sprintf("http://%s/v1/livestreams?userId=%s", addr, userId)
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

	var userLivestreams []livestreamgateway.GetLivestreamResponseDTO
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&userLivestreams); err != nil {
		return nil, &gateway.ErrorResponse{
			Message:    fmt.Sprintf("failed to decode resp body: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return userLivestreams, nil
}

func (g *livestreamGateway) CheckIsUserLivestreaming(ctx context.Context, userId string) (bool, *gateway.ErrorResponse) {
	addr, err := g.registry.ServiceAddress(ctx, "livestream")
	if err != nil {
		return false, &gateway.ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusBadGateway,
		}
	}

	url := fmt.Sprintf("http://%s/v1/is-streaming?userId=%s", addr, userId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, &gateway.ErrorResponse{
			Message:    fmt.Sprintf("failed to create the request: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, &gateway.ErrorResponse{
			Message:    fmt.Sprintf("failed to call request: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		resInfo := gateway.ErrorResponse{}
		if err := json.NewDecoder(resp.Body).Decode(&resInfo); err != nil {
			return false, &gateway.ErrorResponse{
				Message:    fmt.Sprintf("failed to decode error response: %s", err),
				StatusCode: http.StatusInternalServerError,
			}
		}

		return false, &resInfo
	}

	defer resp.Body.Close()
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return false, &gateway.ErrorResponse{
			Message:    fmt.Sprintf("failed to parse error response: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	res, err := strconv.ParseBool(buf.String())
	if err != nil {
		return false, &gateway.ErrorResponse{
			Message:    fmt.Sprintf("failed to parse error response: %s", err),
			StatusCode: http.StatusInternalServerError,
		}
	}

	return res, nil
}
