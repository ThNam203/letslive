package userservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"sen1or/letslive/chat/domains"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

const requestTimeout = 3 * time.Second

type Identity struct {
	ID             string  `json:"id"`
	Username       string  `json:"username"`
	ProfilePicture *string `json:"profilePicture"`
}

// Gateway resolves user identities through the user service. Display names
// are never taken from a caller's payload: the user service owns the
// username, so it is the only trusted source.
type Gateway struct {
	registry discovery.Registry
	client   *http.Client
}

func NewGateway(registry discovery.Registry) *Gateway {
	return &Gateway{registry: registry, client: &http.Client{Timeout: requestTimeout}}
}

type batchRequest struct {
	IDs []string `json:"ids"`
}

type batchResponse struct {
	Data []Identity `json:"data"`
}

// GetIdentities returns the identities keyed by user id. Unknown ids are
// simply absent. Any failure to reach the user service is ErrUserService.
func (g *Gateway) GetIdentities(ctx context.Context, ids []string) (map[string]Identity, error) {
	unique := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return map[string]Identity{}, nil
	}

	identities, err := g.fetch(ctx, unique)
	if err != nil {
		logger.Errorf(ctx, "failed to resolve user identities: %v", err)
		return nil, fmt.Errorf("%w: %v", domains.ErrUserService, err)
	}
	return identities, nil
}

func (g *Gateway) fetch(ctx context.Context, ids []string) (map[string]Identity, error) {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		return nil, fmt.Errorf("resolve user service: %w", err)
	}

	body, err := json.Marshal(batchRequest{IDs: ids})
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/v1/internal/users/batch", addr), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", requestID(ctx))

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call user service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("user service returned %d", resp.StatusCode)
	}

	var decoded batchResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	identities := make(map[string]Identity, len(decoded.Data))
	for _, identity := range decoded.Data {
		identities[identity.ID] = identity
	}
	return identities, nil
}

func requestID(ctx context.Context) string {
	if id, ok := ctx.Value("requestId").(string); ok && id != "" {
		return id
	}
	id, err := uuid.NewV4()
	if err != nil {
		return uuid.Nil.String()
	}
	return id.String()
}
