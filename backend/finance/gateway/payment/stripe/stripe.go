package stripe

import (
	"context"
	"fmt"
	"sen1or/letslive/finance/domains"
	gatewaypayment "sen1or/letslive/finance/gateway/payment"

	stripego "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/webhook"
)

type StripeGateway struct {
	apiKey        string
	webhookSecret string
	successURL    string
	cancelURL     string
}

func NewStripeGateway(apiKey, webhookSecret, successURL, cancelURL string) *StripeGateway {
	return &StripeGateway{
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
		successURL:    successURL,
		cancelURL:     cancelURL,
	}
}

func (g *StripeGateway) Provider() domains.PaymentProvider {
	return domains.PaymentProviderStripe
}

func (g *StripeGateway) CreateCheckoutSession(ctx context.Context, idempotencyKey string, amount int64, currencyCode string, metadata map[string]string) (*gatewaypayment.CheckoutSession, error) {
	_ = ctx

	stripego.Key = g.apiKey

	params := &stripego.CheckoutSessionParams{
		Mode:       stripego.String(string(stripego.CheckoutSessionModePayment)),
		SuccessURL: stripego.String(g.successURL),
		CancelURL:  stripego.String(g.cancelURL),
		LineItems: []*stripego.CheckoutSessionLineItemParams{
			{
				PriceData: &stripego.CheckoutSessionLineItemPriceDataParams{
					Currency: stripego.String(currencyCode),
					ProductData: &stripego.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripego.String(fmt.Sprintf("LetsLive %s deposit", currencyCode)),
					},
					UnitAmount: stripego.Int64(amount),
				},
				Quantity: stripego.Int64(1),
			},
		},
	}
	if metadata != nil {
		params.Metadata = metadata
	}
	params.SetIdempotencyKey(idempotencyKey)

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe checkout session: %w", err)
	}
	return &gatewaypayment.CheckoutSession{
		ProviderRef: sess.ID,
		CheckoutURL: sess.URL,
	}, nil
}

func (g *StripeGateway) VerifyWebhook(payload []byte, signature string) (*gatewaypayment.WebhookEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, g.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("stripe webhook signature: %w", err)
	}

	switch event.Type {
	case "checkout.session.completed":
		var sess stripego.CheckoutSession
		if err := unmarshalDataObject(event, &sess); err != nil {
			return nil, err
		}
		return &gatewaypayment.WebhookEvent{Type: gatewaypayment.WebhookEventCompleted, ProviderRef: sess.ID}, nil
	case "checkout.session.async_payment_failed", "checkout.session.expired":
		var sess stripego.CheckoutSession
		if err := unmarshalDataObject(event, &sess); err != nil {
			return nil, err
		}
		return &gatewaypayment.WebhookEvent{Type: gatewaypayment.WebhookEventFailed, ProviderRef: sess.ID}, nil
	}
	return &gatewaypayment.WebhookEvent{Type: gatewaypayment.WebhookEventIgnored}, nil
}
