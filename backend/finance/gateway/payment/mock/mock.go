package mock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sen1or/letslive/finance/domains"
	gatewaypayment "sen1or/letslive/finance/gateway/payment"
)

// MockGateway returns deterministic fake checkout URLs and accepts any signed
// payload of the form {"type":"completed|failed","providerRef":"..."}. Used
// for local development before real Stripe wiring lands.
type MockGateway struct{}

func NewMockGateway() *MockGateway {
	return &MockGateway{}
}

func (g *MockGateway) Provider() domains.PaymentProvider {
	return domains.PaymentProvider("mock")
}

func (g *MockGateway) CreateCheckoutSession(ctx context.Context, idempotencyKey string, amount int64, currencyCode string, metadata map[string]string) (*gatewaypayment.CheckoutSession, error) {
	_ = metadata
	providerRef := "mock_" + idempotencyKey
	return &gatewaypayment.CheckoutSession{
		ProviderRef: providerRef,
		CheckoutURL: fmt.Sprintf("mock://checkout/%s?amount=%d&currency=%s", providerRef, amount, currencyCode),
	}, nil
}

func (g *MockGateway) VerifyWebhook(payload []byte, signature string) (*gatewaypayment.WebhookEvent, error) {
	_ = signature
	var body struct {
		Type        string `json:"type"`
		ProviderRef string `json:"providerRef"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil, fmt.Errorf("mock webhook: invalid payload: %w", err)
	}

	switch body.Type {
	case string(gatewaypayment.WebhookEventCompleted):
		return &gatewaypayment.WebhookEvent{Type: gatewaypayment.WebhookEventCompleted, ProviderRef: body.ProviderRef}, nil
	case string(gatewaypayment.WebhookEventFailed):
		return &gatewaypayment.WebhookEvent{Type: gatewaypayment.WebhookEventFailed, ProviderRef: body.ProviderRef}, nil
	}
	return nil, errors.New("mock webhook: unknown event type")
}
