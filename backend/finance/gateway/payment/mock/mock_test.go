package mock

import (
	"context"
	"sen1or/letslive/finance/domains"
	gatewaypayment "sen1or/letslive/finance/gateway/payment"
	"strings"
	"testing"
)

func TestMockProvider(t *testing.T) {
	g := NewMockGateway()
	if g.Provider() != domains.PaymentProvider("mock") {
		t.Fatalf("expected provider 'mock', got %q", g.Provider())
	}
}

func TestMockCreateCheckoutSession(t *testing.T) {
	g := NewMockGateway()
	sess, err := g.CreateCheckoutSession(context.Background(), "key-123", 5000, "SPARK", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.ProviderRef != "mock_key-123" {
		t.Errorf("provider ref = %q, want 'mock_key-123'", sess.ProviderRef)
	}
	if !strings.HasPrefix(sess.CheckoutURL, "mock://checkout/mock_key-123") {
		t.Errorf("checkout url has wrong prefix: %q", sess.CheckoutURL)
	}
}

func TestMockVerifyWebhookCompleted(t *testing.T) {
	g := NewMockGateway()
	event, err := g.VerifyWebhook([]byte(`{"type":"completed","providerRef":"mock_abc"}`), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Type != gatewaypayment.WebhookEventCompleted {
		t.Errorf("event type = %q, want completed", event.Type)
	}
	if event.ProviderRef != "mock_abc" {
		t.Errorf("provider ref = %q, want 'mock_abc'", event.ProviderRef)
	}
}

func TestMockVerifyWebhookFailed(t *testing.T) {
	g := NewMockGateway()
	event, err := g.VerifyWebhook([]byte(`{"type":"failed","providerRef":"mock_xyz"}`), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Type != gatewaypayment.WebhookEventFailed {
		t.Errorf("event type = %q, want failed", event.Type)
	}
}

func TestMockVerifyWebhookRejects(t *testing.T) {
	g := NewMockGateway()
	if _, err := g.VerifyWebhook([]byte(`not json`), ""); err == nil {
		t.Errorf("expected error for invalid JSON")
	}
	if _, err := g.VerifyWebhook([]byte(`{"type":"unknown"}`), ""); err == nil {
		t.Errorf("expected error for unknown event type")
	}
}
