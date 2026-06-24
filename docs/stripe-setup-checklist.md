# Stripe Deposit Setup Checklist

## 1 — Get Stripe credentials

1. Go to https://dashboard.stripe.com/apikeys
2. Copy **Secret key** (`sk_test_...` for test, `sk_live_...` for prod)
3. Go to https://dashboard.stripe.com/webhooks → **Add endpoint**
   - URL: `https://<your-domain>/deposits/webhook/stripe`
   - Events to select:
     - `checkout.session.completed`
     - `checkout.session.async_payment_failed`
     - `checkout.session.expired`
   - After creating: copy the **Signing secret** (`whsec_...`)

## 2 — Set environment variables

In your `.env` (or server environment):

```env
FINANCE_STRIPE_API_KEY=sk_test_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
FINANCE_STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

## 3 — Set config server YAML (finance service config)

Add/update the `stripe` block in the finance service YAML in your config server Git repo:

```yaml
stripe:
  successUrl: "https://<your-frontend>/wallet/overview"
  cancelUrl: "https://<your-frontend>/wallet/deposit"
  fiatCurrencyCode: "usd"
```

> `fiatCurrencyCode` must be a valid ISO 4217 code that Stripe accepts (usd, eur, sgd, etc.)
> SPARK and FLARE are virtual — this is the real currency Stripe charges.
> Exchange rate for MVP: 1 SPARK/FLARE minor unit = 1 fiat cent (100 SPARK = $1.00 USD).

## 4 — Local development (no public URL)

Use Stripe CLI to forward webhooks to your local server:

```sh
stripe login
stripe listen --forward-to http://localhost:8000/deposits/webhook/stripe
```

The CLI prints a webhook secret — use that as `FINANCE_STRIPE_WEBHOOK_SECRET` for local dev
(it's different from the dashboard secret).

## 5 — Verify finance service is running

```sh
curl http://localhost:8000/currencies
# Should return: {"success":true,"data":[{"code":"FLARE",...},{"code":"SPARK",...}]}
```

## 6 — Test a deposit

```sh
curl -X POST http://localhost:8000/deposits \
  -H "Content-Type: application/json" \
  -H "Cookie: ACCESS_TOKEN=<your-jwt>" \
  -d '{"provider":"stripe","currencyCode":"SPARK","amount":"10.00"}'
# Returns checkoutUrl — open it in browser to complete payment
```

After completing payment in Stripe's test UI, check wallet balance:

```sh
curl http://localhost:8000/wallet \
  -H "Cookie: ACCESS_TOKEN=<your-jwt>"
# balances should show SPARK credited
```

## Test cards (Stripe test mode)

| Card | Result |
|------|--------|
| `4242 4242 4242 4242` | Success |
| `4000 0000 0000 9995` | Insufficient funds (failure) |

Expiry: any future date. CVC: any 3 digits.

---

## What was fixed (already applied to code)

| Issue | Fix |
|-------|-----|
| Stripe gateway sent `"SPARK"` as currency (invalid) | Gateway now uses `fiatCurrencyCode` config |
| Frontend API paths had `/finance/` prefix (Kong has no such route) | Removed — now `/wallet`, `/deposits`, etc. |
| `PaymentStatus.PENDING` didn't match backend `"created"` | Renamed to `PaymentStatus.CREATED` |
