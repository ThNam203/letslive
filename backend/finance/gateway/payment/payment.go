package payment

import (
	"context"
	"sen1or/letslive/finance/domains"
)

// CheckoutSession is the payment-gateway-agnostic checkout result returned to
// the frontend.
type CheckoutSession struct {
	ProviderRef string
	CheckoutURL string
}

// WebhookEventType abstracts the provider-specific event names.
type WebhookEventType string

const (
	WebhookEventCompleted WebhookEventType = "completed"
	WebhookEventFailed    WebhookEventType = "failed"
	WebhookEventIgnored   WebhookEventType = "ignored"
)

type WebhookEvent struct {
	Type        WebhookEventType
	ProviderRef string
}

// PaymentGateway abstracts the upstream provider (Stripe, PayPal, mock).
type PaymentGateway interface {
	Provider() domains.PaymentProvider
	CreateCheckoutSession(ctx context.Context, idempotencyKey string, amount int64, currencyCode string, metadata map[string]string) (*CheckoutSession, error)
	VerifyWebhook(payload []byte, signature string) (*WebhookEvent, error)
}
