package contenthttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/gateway"
	contentgateway "sen1or/letslive/user/gateway/content"

	"github.com/gofrs/uuid/v5"
)

var contentServices = []string{"vod", "livestream"}

type contentHTTPGateway struct {
	registry discovery.Registry
}

func NewContentGateway(registry discovery.Registry) contentgateway.ContentGateway {
	return &contentHTTPGateway{
		registry: registry,
	}
}

func (g *contentHTTPGateway) SetAuthorDisabled(ctx context.Context, userId uuid.UUID, disabled bool) error {
	body := struct {
		Disabled bool `json:"disabled"`
	}{Disabled: disabled}

	return g.putToAll(ctx, "/v1/internal/disabled-authors/"+userId.String(), body)
}

func (g *contentHTTPGateway) ReplaceDisabledAuthors(ctx context.Context, userIds []uuid.UUID) error {
	if userIds == nil {
		userIds = []uuid.UUID{}
	}
	body := struct {
		UserIds []uuid.UUID `json:"userIds"`
	}{UserIds: userIds}

	return g.putToAll(ctx, "/v1/internal/disabled-authors", body)
}

func (g *contentHTTPGateway) putToAll(ctx context.Context, path string, body any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	for _, service := range contentServices {
		if err := g.put(ctx, service, path, payload); err != nil {
			return err
		}
	}

	return nil
}

func (g *contentHTTPGateway) put(ctx context.Context, service string, path string, payload []byte) error {
	addr, err := g.registry.ServiceAddress(ctx, service)
	if err != nil {
		logger.Errorf(ctx, "failed to get %s service address: %v", service, err)
		return fmt.Errorf("%s service unavailable", service)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("http://%s%s", addr, path), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Warnf(ctx, "failed to set request id header: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(ctx, "failed to call %s service %s: %v", service, path, err)
		return fmt.Errorf("failed to call %s service", service)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("%s service returned status %d on %s", service, resp.StatusCode, path)
	}

	return nil
}
