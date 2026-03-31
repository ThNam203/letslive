# Finance Service - Backend API Specification

This document describes all the API endpoints that the **web** and **mobile** frontends expect from the finance service. Use this as a reference when implementing the backend.

All endpoints are exposed through Kong API gateway under the `/finance/` prefix.
All endpoints require authentication (user extracted from JWT cookie/token).
All responses follow the shared `Response[T]` format:

```json
{
  "requestId": "string",
  "success": true,
  "code": 100000,
  "key": "res_succ_ok",
  "message": "Success",
  "data": {},
  "meta": { "page": 0, "page_size": 20, "total": 100 },
  "errorDetails": null
}
```

---

## Error Codes (600xx range)

| Code  | Key                            | Description                     |
|-------|--------------------------------|---------------------------------|
| 60000 | `res_err_account_not_found`    | User wallet not found           |
| 60001 | `res_err_account_frozen`       | Account is frozen               |
| 60002 | `res_err_insufficient_balance` | Not enough balance              |
| 60003 | `res_err_invalid_amount`       | Amount is <= 0 or not a number  |
| 60004 | `res_err_transaction_failed`   | Transaction could not complete  |
| 60005 | `res_err_payment_failed`       | Payment provider returned error |
| 60006 | `res_err_payment_not_found`    | Payment record not found        |
| 60007 | `res_err_unsupported_currency` | Currency code not recognized    |
| 60008 | `res_err_deposit_limit_exceeded`| Exceeded max deposit amount    |

---

## 1. GET `/finance/wallet`

**Purpose:** Get the current user's wallet overview (account + all balances).

**Auth:** Required. Extract `userId` from JWT.

**Logic:**
1. Look up the `accounts` table for `owner_id = userId` and `type = 'user_wallet'`
2. If no account exists, **auto-create** one with `status = 'active'`
3. Fetch all rows from `account_balances` for this account
4. If balances don't exist for SPARK/FLARE, return `"0"` for those currencies

**Response:**
```json
{
  "data": {
    "account": {
      "id": "uuid",
      "ownerId": "user-uuid",
      "type": "user_wallet",
      "status": "active",
      "createdAt": "2026-03-31T12:00:00Z",
      "updatedAt": "2026-03-31T12:00:00Z"
    },
    "balances": [
      {
        "accountId": "uuid",
        "currencyCode": "SPARK",
        "balance": "1500.00",
        "lastEntryId": "uuid-or-null"
      },
      {
        "accountId": "uuid",
        "currencyCode": "FLARE",
        "balance": "250.00",
        "lastEntryId": "uuid-or-null"
      }
    ]
  }
}
```

**Errors:** `60000` if account lookup fails unexpectedly, `60001` if account is frozen.

---

## 2. GET `/finance/currencies`

**Purpose:** List all supported currencies.

**Auth:** Not strictly required, but keeping it authenticated is fine.

**Logic:** Query `currencies` table. Return all rows.

**Response:**
```json
{
  "data": [
    { "code": "SPARK", "name": "Spark", "precision": 2 },
    { "code": "FLARE", "name": "Flare", "precision": 2 }
  ]
}
```

---

## 3. GET `/finance/transactions?page=0&page_size=20`

**Purpose:** List the current user's transactions, newest first.

**Auth:** Required. Extract `userId` from JWT.

**Logic:**
1. Get user's account (same as wallet endpoint)
2. Query `transactions` where `actor_id = userId` ORDER BY `created_at DESC`
3. For each transaction, optionally join `ledger_entries` WHERE `account_id = user's account id`
   - Only return entries **relevant to the user's account**, not all entries in the transaction
4. Apply pagination using `LIMIT` and `OFFSET`
5. Return total count in `meta`

**Query params:**
- `page` (int, default 0)
- `page_size` (int, default 20, max 100)

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "type": "donate",
      "status": "completed",
      "reference": "donation-123",
      "description": "Donation to streamer",
      "actorId": "user-uuid",
      "metadata": null,
      "createdAt": "2026-03-30T10:00:00Z",
      "updatedAt": "2026-03-30T10:00:01Z",
      "entries": [
        {
          "id": "uuid",
          "transactionId": "uuid",
          "accountId": "user-account-uuid",
          "currencyCode": "SPARK",
          "amount": "-100.00",
          "createdAt": "2026-03-30T10:00:00Z"
        }
      ]
    }
  ],
  "meta": { "page": 0, "page_size": 20, "total": 42 }
}
```

**Note on `entries`:** The `amount` field is a signed decimal string:
- Negative = money left the user's account (debit)
- Positive = money entered the user's account (credit)

---

## 4. GET `/finance/transactions/:id`

**Purpose:** Get a single transaction with its ledger entries.

**Auth:** Required. Verify the transaction belongs to the user.

**Response:** Same shape as a single item from the list endpoint.

**Errors:** `60004` if not found or not owned by user.

---

## 5. POST `/finance/deposits`

**Purpose:** Initiate a deposit (purchase virtual currency with real money).

**Auth:** Required.

**Request body:**
```json
{
  "provider": "stripe",
  "currencyCode": "SPARK",
  "amount": "50.00"
}
```

**Validation:**
- `provider` must be `"stripe"` or `"paypal"`
- `currencyCode` must be `"SPARK"` or `"FLARE"`
- `amount` must be a valid decimal > 0, min 1, max 10000

**Logic:**
1. Validate input
2. Create a `payments` record with `status = 'pending'`
3. Call the payment provider API (Stripe Checkout Session / PayPal Order):
   - Create a checkout session with the amount
   - Store the `provider_reference` (e.g., Stripe session ID)
   - Generate a `checkoutUrl` for the frontend to redirect to
4. Return the payment record + checkout URL

**Response:**
```json
{
  "data": {
    "payment": {
      "id": "uuid",
      "transactionId": null,
      "provider": "stripe",
      "providerReference": "cs_live_abc123",
      "currencyCode": "SPARK",
      "amount": "50.00",
      "status": "pending",
      "createdAt": "2026-03-31T12:00:00Z",
      "updatedAt": "2026-03-31T12:00:00Z"
    },
    "checkoutUrl": "https://checkout.stripe.com/c/pay/cs_live_abc123"
  }
}
```

**Errors:** `60003`, `60007`, `60008`

---

## 6. POST `/finance/deposits/webhook` (Internal - not called by frontend)

**Purpose:** Webhook handler for Stripe/PayPal payment confirmations.

**Auth:** Signature verification from payment provider (not JWT).

**Logic:**
1. Verify webhook signature (Stripe: `stripe-signature` header)
2. Find the `payments` record by `provider_reference`
3. If payment event = `completed`:
   a. Update payment `status = 'completed'`
   b. Create a `transaction` with `type = 'purchase'`, `status = 'completed'`
   c. Create ledger entries:
      - Credit user's account: `+amount` in the chosen currency
      - Debit platform's escrow/source account: `-amount`
   d. Update `account_balances` for the user
   e. Link the transaction to the payment record
4. If payment event = `failed`:
   a. Update payment `status = 'failed'`

**Note:** This is NOT called by the frontend but is critical for the deposit flow to complete. The webhook URL is registered with Stripe/PayPal during setup.

---

## 7. GET `/finance/payments?page=0&page_size=20`

**Purpose:** List the current user's payment history (deposits).

**Auth:** Required.

**Logic:**
1. Get user's account
2. Query `payments` table joined through `transactions` where `actor_id = userId`
   - Alternatively, store `user_id` on payments directly for easier querying
3. Order by `created_at DESC`, apply pagination

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "transactionId": "uuid-or-null",
      "provider": "stripe",
      "providerReference": "cs_live_abc123",
      "currencyCode": "SPARK",
      "amount": "50.00",
      "status": "completed",
      "createdAt": "2026-03-31T12:00:00Z",
      "updatedAt": "2026-03-31T12:01:00Z"
    }
  ],
  "meta": { "page": 0, "page_size": 20, "total": 5 }
}
```

---

## 8. GET `/finance/payments/:id`

**Purpose:** Get a single payment record.

**Auth:** Required. Verify ownership.

**Response:** Same shape as a single item from the list endpoint.

**Errors:** `60006` if not found.

---

## Implementation Order (Recommended)

1. **Models/DTOs** - Define Go structs for Account, Balance, Transaction, LedgerEntry, Payment
2. **Repository layer** - CRUD operations for each table
3. **GET /finance/wallet** - The simplest endpoint, auto-create account
4. **GET /finance/currencies** - Static data query
5. **GET /finance/transactions** - List with pagination
6. **GET /finance/transactions/:id** - Single item
7. **POST /finance/deposits** - Create payment + mock checkout URL (before integrating real Stripe)
8. **GET /finance/payments** - List payments
9. **Stripe/PayPal integration** - Real checkout sessions + webhook handler
10. **POST /finance/deposits/webhook** - Complete the payment flow

## Tips for Implementation

- Use the **shared response package** (`sen1or/letslive/shared/response`) for all responses
- Use the **shared config** pattern with `PostProcess` for DB credentials
- Register routes in `api/http.go` following the existing `wrap()` pattern
- For the double-entry bookkeeping, always create ledger entries inside a **database transaction** (BEGIN/COMMIT) to maintain the zero-sum invariant
- The `account_balances` table is a **denormalized cache** - update it atomically alongside ledger entries
- For Stripe integration, use the official Go SDK: `github.com/stripe/stripe-go/v82`
- For development/testing, you can mock the checkout URL (return a fake URL) and manually trigger the webhook logic

## Kong Gateway Configuration

Register the finance service in Kong to route `/finance/*` to the finance backend:

```
Service name: finance-service
Service host: finance.service.consul (via Consul DNS)
Route path: /finance
Strip path: true
```

This means the finance backend receives requests at `/wallet`, `/currencies`, etc. (without the `/finance` prefix), which matches the routes defined in `api/http.go`.
