# Shop & Gift Feature Design

**Date:** 2026-06-24  
**Branch:** feat/finance-service  
**Author:** ThNam203

---

## Overview

Users can browse a shop to buy virtual gift items using platform currency, then send those gifts to other users from anywhere in the app (profile pages, search results, etc.). Gifting triggers a visual animation for the sender and adds the item to the recipient's public collection. Distinct from the donate system (direct currency transfer).

---

## Decisions Made

| Question | Decision |
|---|---|
| Gifting context | Anywhere — any user profile/search |
| Currency flow | Gifts are shop items; donate system handles direct currency transfer |
| Gift effect | Animation (sender-side) + collectible added to recipient's collection |
| Catalog management | Admin-managed via DB for MVP; admin panel later |
| Purchase flow | Both: buy-and-send immediately OR buy to inventory and send later |
| Architecture | Approach C: shop catalog in finance service, inventory + gifting in user service |

---

## Data Model

### Finance service — new table

```sql
CREATE TABLE shop_items (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name          VARCHAR(100) NOT NULL,
  description   TEXT,
  image_url     TEXT NOT NULL,
  animation_url TEXT NOT NULL,
  price         INTEGER NOT NULL,   -- in platform currency units
  is_active     BOOLEAN NOT NULL DEFAULT true,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### User service — two new tables

```sql
-- Items a user currently owns
CREATE TABLE user_inventory (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  shop_item_id  UUID NOT NULL,       -- cross-service ref, no FK
  quantity      INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, shop_item_id)
);

-- Every gift-send event; recipient's collection is derived from this
CREATE TABLE gifts (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  sender_user_id    UUID NOT NULL REFERENCES users(id),
  recipient_user_id UUID NOT NULL REFERENCES users(id),
  shop_item_id      UUID NOT NULL,   -- cross-service ref, no FK
  quantity          INTEGER NOT NULL DEFAULT 1 CHECK (quantity >= 1),
  message           TEXT,
  sent_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

No FK across service boundaries. Item metadata fetched from finance service when rendering gift history.

---

## API Design

### Finance service — new endpoints

```
GET  /v1/shop/items              # list active items (public)
GET  /v1/shop/items/{id}         # single item detail (public)
POST /v1/shop/purchase           # buy items, deduct wallet (private/JWT)
```

`POST /v1/shop/purchase` request body:
```json
{
  "shop_item_id": "uuid",
  "quantity": 1,
  "recipient_user_id": "uuid"   // optional — present = quick-send flow
}
```

Response includes `animation_url` so client can play the animation immediately on quick-send.

When `recipient_user_id` is present with `quantity > 1`, all N items are sent as a bulk gift — a single `gifts` record with the quantity stored, not N separate records.

### User service — new endpoints

```
GET  /v1/user/me/inventory                 # my owned items (private)
POST /v1/gifts                             # send from inventory (private)
GET  /v1/user/{userId}/gifts/received      # recipient's public collection
GET  /v1/user/me/gifts/sent                # my sent gift history (private)

# Internal — service-to-service only, no JWT, internal network
POST /v1/internal/inventory/add            # finance → user after purchase
POST /v1/internal/gifts/create             # finance → user for quick-send
```

`POST /v1/gifts` body:
```json
{
  "shop_item_id": "uuid",
  "recipient_user_id": "uuid",
  "message": "optional note"
}
```

---

## Purchase Flows

### Quick-send (buy + send in one action)

```
Client
  → POST /finance/v1/shop/purchase { shop_item_id, quantity: 1, recipient_user_id }
Finance service
  → Validate wallet balance ≥ item price
  → Create transaction (type: purchase) + ledger entries (atomic)
  → POST /user/v1/internal/gifts/create { sender_id, recipient_id, shop_item_id, message? }
User service
  → Insert gifts record
  → CreateNotification (type: gift_received, reference_id: gift.id)
Response
  → { gift_id, animation_url }   ← client plays animation
```

### Buy to inventory, send later

```
Client
  → POST /finance/v1/shop/purchase { shop_item_id, quantity: N }
Finance service
  → Validate balance, create transaction + ledger entries
  → POST /user/v1/internal/inventory/add { user_id, shop_item_id, quantity: N }
User service
  → UPSERT user_inventory (quantity += N)

Later:
Client
  → POST /user/v1/gifts { shop_item_id, recipient_user_id, message? }
User service
  → Deduct inventory (quantity -= 1, CHECK quantity >= 0)
  → Insert gifts record
  → CreateNotification (type: gift_received, reference_id: gift.id)
```

---

## Backend Code Structure

Follows existing conventions: one file per operation, `pgx.RowToStructByNameLax`, `BaseHandler` embed, `wrap()` route registration, tracer spans per handler.

### Finance service — new files

```
domains/
  shop_item.go                    # ShopItem struct + ShopItemRepository interface

repositories/shop_item/
  shop_item.go                    # postgresShopItemRepo + constructor
  list.go                         # List active items
  get_by_id.go                    # Get single item

services/shop_item/
  shop_item.go                    # ShopItemService{repo}

services/purchase/
  purchase.go                     # PurchaseService{walletRepo, txRepo, userServiceClient}

handlers/shop_item/
  shop_item.go                    # ShopItemHandler embeds BaseHandler
  get_items_public.go             # GET /v1/shop/items
  get_item_public.go              # GET /v1/shop/items/{id}

handlers/purchase/
  purchase.go                     # PurchaseHandler embeds BaseHandler
  create_purchase_private.go      # POST /v1/shop/purchase

dto/
  shop_item.go                    # ShopItemResponseDTO
  purchase.go                     # PurchaseRequestDTO, PurchaseResponseDTO
```

### User service — new files

```
domains/
  inventory.go                    # UserInventory struct + InventoryRepository interface
  gift.go                         # Gift struct + GiftRepository interface

repositories/inventory/
  inventory.go                    # constructor
  upsert.go                       # INSERT ... ON CONFLICT DO UPDATE (add quantity)
  deduct.go                       # quantity-- with quantity >= 0 guard
  get_by_user_id.go               # paginated list

repositories/gift/
  gift.go                         # constructor
  create.go                       # insert gift record
  list_by_recipient.go            # GET collection (paginated)
  list_by_sender.go               # GET sent history (paginated)

services/
  inventory.go                    # InventoryService{inventoryRepo}
  gift.go                         # GiftService{giftRepo, inventoryRepo, notificationService}

handlers/inventory/
  inventory.go                    # InventoryHandler embeds BaseHandler
  get_inventory_private.go        # GET /v1/user/me/inventory

handlers/gift/
  gift.go                         # GiftHandler embeds BaseHandler
  send_gift_private.go            # POST /v1/gifts
  get_gifts_received_public.go    # GET /v1/user/{userId}/gifts/received
  get_gifts_sent_private.go       # GET /v1/user/me/gifts/sent
  add_inventory_internal.go       # POST /v1/internal/inventory/add
  create_gift_internal.go         # POST /v1/internal/gifts/create
```

**Inter-service call:** Finance calls User service via internal Docker network hostname (Consul service name), not through Kong. Finance service config gains a `UserServiceBaseURL` field.

---

## Mobile UI (Flutter)

```
features/shop/
  presentation/
    shop_screen.dart              # grid of items, price badge, buy/send CTA
    shop_item_detail.dart         # item detail bottom sheet
  data/
    shop_repository.dart          # GET /finance/shop/items

features/gift/
  presentation/
    gift_picker_sheet.dart        # bottom sheet on profile: inventory + quick-buy
    gift_animation_overlay.dart   # plays animation_url on send confirmation
    gifts_received_screen.dart    # public collection tab on profile
  data/
    gift_repository.dart          # POST /user/gifts, GET gifts/received

features/wallet/ (extend existing)
  presentation/
    inventory_tab.dart            # new tab: owned items + quantities
```

New routes in `app_router.dart`:
```dart
static const shop = '/shop';
static const giftsReceived = '/user/:id/gifts';
```

---

## Web UI (Next.js)

```
app/[lng]/(main)/
  shop/
    page.tsx                      # shop catalog grid
  user/[id]/
    gifts/page.tsx                # public gifts collection (profile tab)
  (account)/
    inventory/page.tsx            # my owned items
```

Gift button added to existing user profile page — opens gift picker modal. On send confirmation, animation plays client-side using `animation_url`.

---

## Notification Type

New notification type added to user service notification domain:

```go
const NotificationTypeGiftReceived = "gift_received"
```

Notification payload:
- `title`: "{sender_name} sent you a gift!"
- `message`: "{item_name}" (+ message if included)
- `reference_id`: gift UUID
- `action_url`: `/user/me/gifts/received`

---

## Out of Scope (MVP)

- Admin panel UI for catalog management (DB seed only)
- Gift animation on recipient's screen in real-time (push/WebSocket)
- Gifting during livestream (separate feature, separate design)
- Withdraw or monetize received gifts
- Gift quantity limits or cooldowns
