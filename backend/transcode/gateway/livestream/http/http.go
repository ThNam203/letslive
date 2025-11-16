package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/letslive/transcode/gateway"
	"sen1or/letslive/transcode/pkg/discovery"
	"sen1or/letslive/transcode/pkg/logger"
	"sen1or/letslive/transcode/response"
)

type LivestreamGateway struct {
	registry discovery.Registry
}

func NewLivestreamGateway(registry discovery.Registry) *LivestreamGateway {
	return &LivestreamGateway{
		registry: registry,
	}
}

func (g *LivestreamGateway) Create(ctx context.Context, data CreateLivestreamRequestDTO) (*LivestreamResponseDTO, *response.Response[any]) {
	addr, err := g.registry.ServiceAddress(ctx, "livestream")
	if err != nil {
		logger.Debugf(ctx, "get service address from gateway failed")
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)

	}

	url := fmt.Sprintf("http://%s/v1/internal/livestreams", addr)
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(data); err != nil {
		logger.Debugf(ctx, "failed to encode data for request: %s", err)
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)

	}
	req, err := http.NewRequest(http.MethodPost, url, payloadBuf)
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
			return nil, &resInfo
		}

		return nil, &resInfo
	}

	var livestreamResponse response.Response[LivestreamResponseDTO]

	if err := json.NewDecoder(resp.Body).Decode(&livestreamResponse); err != nil {
		logger.Debugf(ctx, "failed to decode resp body: %s", err)
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}

	return livestreamResponse.Data, nil
}

func (g *LivestreamGateway) EndLivestream(ctx context.Context, streamId string, endDTO EndLivestreamRequestDTO) *response.Response[any] {
	addr, err := g.registry.ServiceAddress(ctx, "livestream")
	if err != nil {
		logger.Debugf(ctx, "get service address from gateway failed")
		return response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}

	url := fmt.Sprintf("http://%s/v1/internal/livestreams/%s/end", addr, streamId)
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(endDTO); err != nil {
		logger.Debugf(ctx, "failed to encode data for end livestream request: %s", err)
		return response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}

	req, err := http.NewRequest(http.MethodPost, url, payloadBuf)
	if err != nil {
		logger.Debugf(ctx, "failed to create request: %s", err)
		return response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Debugf(ctx, "failed to set request id header: %s", err)
		return response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Debugf(ctx, "failed to call request: %s", err)
		return response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		resInfo := response.Response[any]{}
		if err := json.NewDecoder(resp.Body).Decode(&resInfo); err != nil {
			logger.Debugf(ctx, "failed to decode error response when end user livestream: %s", err)
			return &resInfo
		}

		logger.Debugf(ctx, "failed to end livestream from livestream service with status code: %d", resp.StatusCode)
		return &resInfo
	}

	return nil
}
