# Shop & Gift Web Pages Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build three Next.js pages (shop catalog, wallet inventory, user gifts collection) plus a gift-send modal on user profiles, wired to the existing backend APIs through Kong.

**Architecture:** All pages are client components following the existing pattern — `"use client"`, `fetchClient()` for API calls, Zustand `useUser()` for auth, shadcn/ui + Tailwind 4 for UI. Types live in `types/shop.ts`, API functions in `lib/api/shop.ts` and `lib/api/gift.ts`. The gift modal sits in the user profile directory and is triggered from `profile-header.tsx`.

**Tech Stack:** Next.js 15, React, TypeScript, Tailwind 4, shadcn/ui (Button, Card, Badge, Dialog), fetchClient, Zustand.

## Global Constraints

- All pages: `"use client"` — no server components
- API calls via `fetchClient<ApiResponse<T>>(path)` — path is relative (no `/v1/` prefix; Kong strips it)
- Kong proxy base: `NEXT_PUBLIC_BACKEND_PROTOCOL://NEXT_PUBLIC_BACKEND_IP_ADDRESS:NEXT_PUBLIC_BACKEND_PORT` (resolved by `GLOBAL.API_URL` in `global.ts`)
- Error toasts: `toast.error(t(\`api-response:${res.key}\`), { toastId: res.requestId })`
- Network errors: `toast.error(t("fetch-error:client_fetch_error"))`
- Auth guard: `const user = useUser((s) => s.user); if (!user) return <p>{t(...)}</p>`
- i18n: `useT(["shop", "api-response", "fetch-error"])` pattern; new namespace `shop` in `lib/i18n/locales/{en,vi}/shop.json`
- No new npm dependencies

---

## File Map

| Action | Path | Purpose |
|--------|------|---------|
| Create | `types/shop.ts` | ShopItem, UserInventory, Gift, PurchaseRequest, PurchaseResponse |
| Create | `lib/api/shop.ts` | GetShopItems, GetShopItemById, CreatePurchase |
| Create | `lib/api/gift.ts` | GetMyInventory, GetUserGiftsReceived, SendGift |
| Create | `lib/i18n/locales/en/shop.json` | English strings for shop/inventory/gift pages |
| Create | `lib/i18n/locales/vi/shop.json` | Vietnamese strings |
| Modify | `lib/i18n/locales/en/wallet.json` | Add `navigation.inventory` key |
| Modify | `lib/i18n/locales/vi/wallet.json` | Add `navigation.inventory` key |
| Modify | `app/[lng]/(main)/wallet/layout.tsx` | Add Inventory nav item |
| Create | `app/[lng]/(main)/shop/page.tsx` | Shop catalog grid |
| Create | `app/[lng]/(main)/wallet/inventory/page.tsx` | My owned items |
| Create | `app/[lng]/(main)/users/[userId]/gifts/page.tsx` | Public gifts collection |
| Create | `app/[lng]/(main)/users/[userId]/gift-modal.tsx` | Gift picker dialog |
| Modify | `app/[lng]/(main)/users/[userId]/profile-header.tsx` | Add Gift button that opens modal |

---

## Task 1: Types (`types/shop.ts`)

**Files:**
- Create: `web/types/shop.ts`

**Interfaces:**
- Produces: `ShopItem`, `UserInventory`, `Gift`, `PurchaseRequest`, `PurchaseResponse`, `SendGiftRequest` — used by Tasks 2–6

- [ ] **Step 1: Create `types/shop.ts`**

```typescript
// types/shop.ts

export type ShopItem = {
    id: string;
    name: string;
    description: string | null;
    imageUrl: string;
    animationUrl: string;
    price: number;
    createdAt: string;
};

export type UserInventory = {
    id: string;
    userId: string;
    shopItemId: string;
    quantity: number;
    updatedAt: string;
};

export type Gift = {
    id: string;
    senderUserId: string;
    recipientUserId: string;
    shopItemId: string;
    quantity: number;
    message: string | null;
    sentAt: string;
};

export type PurchaseRequest = {
    shopItemId: string;
    quantity: number;
    recipientUserId?: string;
    message?: string;
};

export type PurchaseResponse = {
    giftId: string | null;
    animationUrl: string;
};

export type SendGiftRequest = {
    shop_item_id: string;
    recipient_user_id: string;
    message?: string;
};
```

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/web && npx tsc --noEmit 2>&1 | head -20
```

Expected: no errors related to `types/shop.ts`.

- [ ] **Step 3: Commit**

```bash
git add web/types/shop.ts
git commit -m "feat(web): add shop/gift/inventory TypeScript types

Refs: docs/superpowers/specs/2026-06-24-shop-gift-design.md"
```

---

## Task 2: API Layer (`lib/api/shop.ts` + `lib/api/gift.ts`)

**Files:**
- Create: `web/lib/api/shop.ts`
- Create: `web/lib/api/gift.ts`

**Interfaces:**
- Consumes: `ShopItem`, `UserInventory`, `Gift`, `PurchaseRequest`, `PurchaseResponse`, `SendGiftRequest` from `types/shop.ts`; `ApiResponse` from `types/fetch-response.ts`; `fetchClient` from `utils/fetchClient.ts`
- Produces:
  - `GetShopItems(): Promise<ApiResponse<ShopItem[]>>`
  - `GetShopItemById(id: string): Promise<ApiResponse<ShopItem>>`
  - `CreatePurchase(data: PurchaseRequest): Promise<ApiResponse<PurchaseResponse>>`
  - `GetMyInventory(page?: number, pageSize?: number): Promise<ApiResponse<UserInventory[]>>`
  - `GetUserGiftsReceived(userId: string, page?: number, pageSize?: number): Promise<ApiResponse<Gift[]>>`
  - `SendGift(data: SendGiftRequest): Promise<ApiResponse<void>>`

- [ ] **Step 1: Create `lib/api/shop.ts`**

```typescript
// lib/api/shop.ts
import { ApiResponse } from "@/types/fetch-response";
import { ShopItem, PurchaseRequest, PurchaseResponse } from "@/types/shop";
import { fetchClient } from "@/utils/fetchClient";

export async function GetShopItems(): Promise<ApiResponse<ShopItem[]>> {
    return fetchClient<ApiResponse<ShopItem[]>>(`/shop/items`);
}

export async function GetShopItemById(id: string): Promise<ApiResponse<ShopItem>> {
    return fetchClient<ApiResponse<ShopItem>>(`/shop/items/${id}`);
}

export async function CreatePurchase(
    data: PurchaseRequest,
): Promise<ApiResponse<PurchaseResponse>> {
    return fetchClient<ApiResponse<PurchaseResponse>>(`/shop/purchase`, {
        method: "POST",
        body: JSON.stringify(data),
    });
}
```

- [ ] **Step 2: Create `lib/api/gift.ts`**

```typescript
// lib/api/gift.ts
import { ApiResponse } from "@/types/fetch-response";
import { Gift, UserInventory, SendGiftRequest } from "@/types/shop";
import { fetchClient } from "@/utils/fetchClient";

export async function GetMyInventory(
    page: number = 0,
    pageSize: number = 20,
): Promise<ApiResponse<UserInventory[]>> {
    return fetchClient<ApiResponse<UserInventory[]>>(
        `/user/me/inventory?page=${page}&page_size=${pageSize}`,
    );
}

export async function GetUserGiftsReceived(
    userId: string,
    page: number = 0,
    pageSize: number = 20,
): Promise<ApiResponse<Gift[]>> {
    return fetchClient<ApiResponse<Gift[]>>(
        `/user/${userId}/gifts/received?page=${page}&page_size=${pageSize}`,
    );
}

export async function SendGift(
    data: SendGiftRequest,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(`/gifts`, {
        method: "POST",
        body: JSON.stringify(data),
    });
}
```

- [ ] **Step 3: Verify TypeScript compiles**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/web && npx tsc --noEmit 2>&1 | head -20
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add web/lib/api/shop.ts web/lib/api/gift.ts
git commit -m "feat(web): add shop and gift API functions

Refs: docs/superpowers/specs/2026-06-24-shop-gift-design.md"
```

---

## Task 3: i18n Strings

**Files:**
- Create: `web/lib/i18n/locales/en/shop.json`
- Create: `web/lib/i18n/locales/vi/shop.json`
- Modify: `web/lib/i18n/locales/en/wallet.json`
- Modify: `web/lib/i18n/locales/vi/wallet.json`

**Interfaces:**
- Produces i18n keys used in Tasks 4–7: `shop:*`, `wallet:navigation.inventory`

- [ ] **Step 1: Create `lib/i18n/locales/en/shop.json`**

```json
{
    "shop": {
        "page_title": "Shop",
        "empty": "No items available right now.",
        "price_label": "{{price}} SPARK",
        "buy_button": "Buy",
        "gift_button": "Send as Gift",
        "gift_send": "Send Gift",
        "gift_sending": "Sending...",
        "gift_sent": "Gift sent!",
        "gift_pick_item": "Pick an item to send",
        "gift_message_placeholder": "Add a message (optional)",
        "gift_no_items": "No items in your inventory.",
        "gift_quick_send": "Quick Buy & Send",
        "purchase_success": "Purchase successful!"
    },
    "inventory": {
        "page_title": "My Inventory",
        "empty": "You don't own any items yet.",
        "quantity_label": "×{{quantity}}",
        "send_gift": "Send as Gift"
    },
    "gifts_received": {
        "page_title": "Gift Collection",
        "empty": "No gifts received yet.",
        "from": "From {{name}}",
        "quantity_label": "×{{quantity}}"
    }
}
```

- [ ] **Step 2: Create `lib/i18n/locales/vi/shop.json`**

```json
{
    "shop": {
        "page_title": "Cửa hàng",
        "empty": "Hiện chưa có sản phẩm nào.",
        "price_label": "{{price}} SPARK",
        "buy_button": "Mua",
        "gift_button": "Tặng quà",
        "gift_send": "Gửi quà",
        "gift_sending": "Đang gửi...",
        "gift_sent": "Đã gửi quà!",
        "gift_pick_item": "Chọn vật phẩm để gửi",
        "gift_message_placeholder": "Thêm tin nhắn (không bắt buộc)",
        "gift_no_items": "Kho đồ của bạn trống.",
        "gift_quick_send": "Mua & Gửi ngay",
        "purchase_success": "Mua thành công!"
    },
    "inventory": {
        "page_title": "Kho đồ của tôi",
        "empty": "Bạn chưa có vật phẩm nào.",
        "quantity_label": "×{{quantity}}",
        "send_gift": "Tặng quà"
    },
    "gifts_received": {
        "page_title": "Bộ sưu tập quà",
        "empty": "Chưa nhận được quà nào.",
        "from": "Từ {{name}}",
        "quantity_label": "×{{quantity}}"
    }
}
```

- [ ] **Step 3: Add `navigation.inventory` to `lib/i18n/locales/en/wallet.json`**

In `web/lib/i18n/locales/en/wallet.json`, find the `"navigation"` object and add the inventory key:

```json
"navigation": {
    "overview": "Overview",
    "transactions": "Transactions",
    "deposit": "Deposit",
    "inventory": "Inventory"
},
```

- [ ] **Step 4: Add `navigation.inventory` to `lib/i18n/locales/vi/wallet.json`**

Open `web/lib/i18n/locales/vi/wallet.json`. In the `"navigation"` object, add:

```json
"inventory": "Kho đồ"
```

- [ ] **Step 5: Commit**

```bash
git add web/lib/i18n/locales/
git commit -m "feat(web): add shop/gift/inventory i18n strings

Refs: docs/superpowers/specs/2026-06-24-shop-gift-design.md"
```

---

## Task 4: Shop Catalog Page

**Files:**
- Create: `web/app/[lng]/(main)/shop/page.tsx`

**Interfaces:**
- Consumes: `GetShopItems()` from `lib/api/shop.ts`; `ShopItem` from `types/shop.ts`
- Produces: `/shop` route rendering a grid of items with name, image, price badge

- [ ] **Step 1: Create `app/[lng]/(main)/shop/page.tsx`**

```tsx
"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import { GetShopItems } from "@/lib/api/shop";
import { ShopItem } from "@/types/shop";
import { Badge } from "@/components/ui/badge";
import IconLoader from "@/components/icons/loader";

export default function ShopPage() {
    const { t } = useT(["shop", "api-response", "fetch-error"]);
    const [items, setItems] = useState<ShopItem[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const fetchItems = async () => {
            setIsLoading(true);
            try {
                const res = await GetShopItems();
                if (res.success && res.data) {
                    setItems(res.data);
                } else {
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                    });
                }
            } catch (_) {
                toast.error(t("fetch-error:client_fetch_error"));
            } finally {
                setIsLoading(false);
            }
        };
        fetchItems();
    }, [t]);

    if (isLoading) {
        return (
            <div className="flex justify-center py-20">
                <IconLoader />
            </div>
        );
    }

    return (
        <div className="p-6">
            <h1 className="text-foreground mb-6 text-3xl font-bold">
                {t("shop:shop.page_title")}
            </h1>

            {items.length === 0 ? (
                <p className="text-muted-foreground text-center py-16">
                    {t("shop:shop.empty")}
                </p>
            ) : (
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
                    {items.map((item) => (
                        <div
                            key={item.id}
                            className="border-border bg-card flex flex-col items-center gap-2 rounded-xl border p-4 transition-shadow hover:shadow-md"
                        >
                            <div className="relative h-24 w-24">
                                <Image
                                    src={item.imageUrl}
                                    alt={item.name}
                                    fill
                                    className="object-contain"
                                    unoptimized
                                />
                            </div>
                            <p className="text-foreground text-center text-sm font-semibold">
                                {item.name}
                            </p>
                            {item.description && (
                                <p className="text-muted-foreground line-clamp-2 text-center text-xs">
                                    {item.description}
                                </p>
                            )}
                            <Badge variant="secondary">
                                {t("shop:shop.price_label", {
                                    price: item.price,
                                })}
                            </Badge>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
```

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/web && npx tsc --noEmit 2>&1 | head -20
```

Expected: no errors.

- [ ] **Step 3: Start dev server and verify page loads**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/web && npm run dev
```

Navigate to `http://localhost:5001/en/shop`. Expected: shop grid renders (or empty state if no items seeded). No console errors.

- [ ] **Step 4: Commit**

```bash
git add "web/app/[lng]/(main)/shop/page.tsx"
git commit -m "feat(web): add shop catalog page

Refs: docs/superpowers/specs/2026-06-24-shop-gift-design.md"
```

---

## Task 5: Wallet Inventory Page

**Files:**
- Create: `web/app/[lng]/(main)/wallet/inventory/page.tsx`
- Modify: `web/app/[lng]/(main)/wallet/layout.tsx`

**Interfaces:**
- Consumes: `GetMyInventory()` from `lib/api/gift.ts`; `UserInventory` from `types/shop.ts`; `useUser` from `hooks/user.ts`
- Produces: `/wallet/inventory` route (auth-gated, inside wallet layout with nav tab)

- [ ] **Step 1: Add Inventory nav item to wallet layout**

In `web/app/[lng]/(main)/wallet/layout.tsx`, find `getNavItems` and add the inventory entry:

```typescript
const getNavItems = (t: any) => [
    { name: t("wallet:navigation.overview"), href: "/wallet/overview" },
    {
        name: t("wallet:navigation.transactions"),
        href: "/wallet/transactions",
    },
    { name: t("wallet:navigation.deposit"), href: "/wallet/deposit" },
    { name: t("wallet:navigation.inventory"), href: "/wallet/inventory" },
];
```

- [ ] **Step 2: Create `app/[lng]/(main)/wallet/inventory/page.tsx`**

```tsx
"use client";

import { useCallback, useEffect, useState } from "react";
import Image from "next/image";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { GetMyInventory } from "@/lib/api/gift";
import { UserInventory } from "@/types/shop";
import { Badge } from "@/components/ui/badge";
import IconLoader from "@/components/icons/loader";

export default function InventoryPage() {
    const { t } = useT(["shop", "api-response", "fetch-error"]);
    const user = useUser((s) => s.user);
    const [items, setItems] = useState<UserInventory[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    const fetchInventory = useCallback(async () => {
        if (!user) return;
        setIsLoading(true);
        try {
            const res = await GetMyInventory();
            if (res.success && res.data) {
                setItems(res.data);
            } else {
                toast.error(t(`api-response:${res.key}`), {
                    toastId: res.requestId,
                });
            }
        } catch (_) {
            toast.error(t("fetch-error:client_fetch_error"));
        } finally {
            setIsLoading(false);
        }
    }, [user, t]);

    useEffect(() => {
        fetchInventory();
    }, [fetchInventory]);

    if (isLoading) {
        return (
            <div className="flex justify-center py-20">
                <IconLoader />
            </div>
        );
    }

    return (
        <>
            <section>
                <h2 className="text-foreground mb-4 text-xl font-semibold">
                    {t("shop:inventory.page_title")}
                </h2>

                {items.length === 0 ? (
                    <p className="text-muted-foreground py-8 text-center text-sm">
                        {t("shop:inventory.empty")}
                    </p>
                ) : (
                    <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4">
                        {items.map((item) => (
                            <div
                                key={item.id}
                                className="border-border bg-card flex flex-col items-center gap-2 rounded-xl border p-4"
                            >
                                <p className="text-muted-foreground text-xs">
                                    {item.shopItemId}
                                </p>
                                <Badge variant="secondary">
                                    {t("shop:inventory.quantity_label", {
                                        quantity: item.quantity,
                                    })}
                                </Badge>
                            </div>
                        ))}
                    </div>
                )}
            </section>
        </>
    );
}
```

> **Note:** `UserInventory` from the backend does not include item metadata (name, image) — it only stores `shopItemId`. The inventory page shows item IDs for now. Enriching with item metadata (a second `GetShopItemById` call per row) is a follow-up if needed; avoid N+1 calls for MVP.

- [ ] **Step 3: Verify TypeScript compiles**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/web && npx tsc --noEmit 2>&1 | head -20
```

Expected: no errors.

- [ ] **Step 4: Open browser and verify**

Navigate to `http://localhost:5001/en/wallet/inventory`. Expected: "Inventory" tab appears in wallet nav. Page renders with inventory items or empty state. Auth guard from `layout.tsx` shows login prompt if not logged in.

- [ ] **Step 5: Commit**

```bash
git add "web/app/[lng]/(main)/wallet/inventory/page.tsx" "web/app/[lng]/(main)/wallet/layout.tsx"
git commit -m "feat(web): add wallet inventory page and nav tab

Refs: docs/superpowers/specs/2026-06-24-shop-gift-design.md"
```

---

## Task 6: User Gifts Collection Page

**Files:**
- Create: `web/app/[lng]/(main)/users/[userId]/gifts/page.tsx`

**Interfaces:**
- Consumes: `GetUserGiftsReceived(userId)` from `lib/api/gift.ts`; `Gift` from `types/shop.ts`; `useParams` from `next/navigation`
- Produces: `/users/[userId]/gifts` public route

- [ ] **Step 1: Create `app/[lng]/(main)/users/[userId]/gifts/page.tsx`**

```tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import { GetUserGiftsReceived } from "@/lib/api/gift";
import { Gift } from "@/types/shop";
import { Badge } from "@/components/ui/badge";
import IconLoader from "@/components/icons/loader";

export default function UserGiftsPage() {
    const { t } = useT(["shop", "api-response", "fetch-error"]);
    const params = useParams<{ userId: string }>();
    const [gifts, setGifts] = useState<Gift[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const fetchGifts = async () => {
            setIsLoading(true);
            try {
                const res = await GetUserGiftsReceived(params.userId);
                if (res.success && res.data) {
                    setGifts(res.data);
                } else {
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                    });
                }
            } catch (_) {
                toast.error(t("fetch-error:client_fetch_error"));
            } finally {
                setIsLoading(false);
            }
        };
        fetchGifts();
    }, [params.userId, t]);

    if (isLoading) {
        return (
            <div className="flex justify-center py-20">
                <IconLoader />
            </div>
        );
    }

    return (
        <div className="p-6">
            <h1 className="text-foreground mb-6 text-3xl font-bold">
                {t("shop:gifts_received.page_title")}
            </h1>

            {gifts.length === 0 ? (
                <p className="text-muted-foreground py-16 text-center">
                    {t("shop:gifts_received.empty")}
                </p>
            ) : (
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
                    {gifts.map((gift) => (
                        <div
                            key={gift.id}
                            className="border-border bg-card flex flex-col items-center gap-2 rounded-xl border p-4"
                        >
                            <p className="text-muted-foreground text-xs">
                                {gift.shopItemId}
                            </p>
                            <Badge variant="secondary">
                                {t("shop:gifts_received.quantity_label", {
                                    quantity: gift.quantity,
                                })}
                            </Badge>
                            {gift.message && (
                                <p className="text-muted-foreground line-clamp-2 text-center text-xs italic">
                                    "{gift.message}"
                                </p>
                            )}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
```

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/web && npx tsc --noEmit 2>&1 | head -20
```

Expected: no errors.

- [ ] **Step 3: Verify page loads in browser**

Navigate to `http://localhost:5001/en/users/<any-user-id>/gifts`. Expected: page renders with gifts or empty state. No console errors.

- [ ] **Step 4: Commit**

```bash
git add "web/app/[lng]/(main)/users/[userId]/gifts/page.tsx"
git commit -m "feat(web): add user gifts collection page

Refs: docs/superpowers/specs/2026-06-24-shop-gift-design.md"
```

---

## Task 7: Gift Button + Modal on User Profile

**Files:**
- Create: `web/app/[lng]/(main)/users/[userId]/gift-modal.tsx`
- Modify: `web/app/[lng]/(main)/users/[userId]/profile-header.tsx`

**Interfaces:**
- Consumes: `GetShopItems()` from `lib/api/shop.ts`; `CreatePurchase()` from `lib/api/shop.ts`; `ShopItem`, `PurchaseRequest` from `types/shop.ts`; `useUser` from `hooks/user.ts`; shadcn `Dialog`, `Button`
- Produces: Gift button on profile header that opens a modal; quick-send flow (`POST /shop/purchase` with `recipientUserId`)

- [ ] **Step 1: Create `gift-modal.tsx`**

```tsx
"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import { GetShopItems, CreatePurchase } from "@/lib/api/shop";
import { ShopItem } from "@/types/shop";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import IconLoader from "@/components/icons/loader";

type GiftModalProps = {
    open: boolean;
    onClose: () => void;
    recipientUserId: string;
    recipientName: string;
};

export default function GiftModal({
    open,
    onClose,
    recipientUserId,
    recipientName,
}: GiftModalProps) {
    const { t } = useT(["shop", "api-response", "fetch-error"]);
    const [items, setItems] = useState<ShopItem[]>([]);
    const [isLoadingItems, setIsLoadingItems] = useState(false);
    const [sendingItemId, setSendingItemId] = useState<string | null>(null);

    useEffect(() => {
        if (!open) return;
        const fetchItems = async () => {
            setIsLoadingItems(true);
            try {
                const res = await GetShopItems();
                if (res.success && res.data) {
                    setItems(res.data);
                } else {
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                    });
                }
            } catch (_) {
                toast.error(t("fetch-error:client_fetch_error"));
            } finally {
                setIsLoadingItems(false);
            }
        };
        fetchItems();
    }, [open, t]);

    const handleSend = async (item: ShopItem) => {
        setSendingItemId(item.id);
        try {
            const res = await CreatePurchase({
                shopItemId: item.id,
                quantity: 1,
                recipientUserId,
            });
            if (res.success) {
                toast.success(t("shop:shop.gift_sent"));
                onClose();
            } else {
                toast.error(t(`api-response:${res.key}`), {
                    toastId: res.requestId,
                });
            }
        } catch (_) {
            toast.error(t("fetch-error:client_fetch_error"));
        } finally {
            setSendingItemId(null);
        }
    };

    return (
        <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
            <DialogContent className="max-w-lg">
                <DialogHeader>
                    <DialogTitle>
                        {t("shop:shop.gift_pick_item")} — {recipientName}
                    </DialogTitle>
                </DialogHeader>

                {isLoadingItems ? (
                    <div className="flex justify-center py-8">
                        <IconLoader />
                    </div>
                ) : items.length === 0 ? (
                    <p className="text-muted-foreground py-8 text-center text-sm">
                        {t("shop:shop.gift_no_items")}
                    </p>
                ) : (
                    <div className="grid grid-cols-3 gap-3 py-2">
                        {items.map((item) => {
                            const isSending = sendingItemId === item.id;
                            return (
                                <button
                                    key={item.id}
                                    onClick={() => handleSend(item)}
                                    disabled={sendingItemId !== null}
                                    className="border-border bg-card hover:border-primary flex flex-col items-center gap-1 rounded-lg border p-3 transition-colors disabled:opacity-50"
                                >
                                    <div className="relative h-16 w-16">
                                        <Image
                                            src={item.imageUrl}
                                            alt={item.name}
                                            fill
                                            className="object-contain"
                                            unoptimized
                                        />
                                    </div>
                                    <p className="text-foreground text-center text-xs font-medium">
                                        {item.name}
                                    </p>
                                    <Badge variant="secondary" className="text-xs">
                                        {t("shop:shop.price_label", {
                                            price: item.price,
                                        })}
                                    </Badge>
                                    {isSending && (
                                        <span className="text-muted-foreground text-xs">
                                            {t("shop:shop.gift_sending")}
                                        </span>
                                    )}
                                </button>
                            );
                        })}
                    </div>
                )}
            </DialogContent>
        </Dialog>
    );
}
```

- [ ] **Step 2: Add Gift button to `profile-header.tsx`**

At the top of `profile-header.tsx`, add the import:

```tsx
import GiftModal from "./gift-modal";
```

Inside the component, add state for modal:

```tsx
const [isGiftModalOpen, setIsGiftModalOpen] = useState(false);
```

After the existing Follow/Unfollow button (inside the `{me?.id && me.id !== user.id && (...)}` block), add the Gift button:

```tsx
{me?.id && me.id !== user.id && (
    <>
        <Button
            variant={user.isFollowing ? "destructive" : "default"}
            disabled={isFetching || !me}
            onClick={onFollowClick}
            className="absolute right-0 bottom-4 flex translate-x-[50%] flex-row items-center justify-center gap-0"
        >
            {isFetching && <IconLoader className="mr-1" />}
            {user.isFollowing ? t("common:unfollow") : t("common:follow")}
        </Button>
        <Button
            variant="outline"
            onClick={() => setIsGiftModalOpen(true)}
            className="absolute right-0 bottom-16 flex translate-x-[50%] flex-row items-center justify-center gap-1"
        >
            🎁 {t("shop:shop.gift_button")}
        </Button>
        <GiftModal
            open={isGiftModalOpen}
            onClose={() => setIsGiftModalOpen(false)}
            recipientUserId={user.id}
            recipientName={user.username}
        />
    </>
)}
```

Also add `"shop"` to the `useT` namespaces array:

```tsx
const { t } = useT([
    "common",
    "users",
    "fetch-error",
    "api-response",
    "accessibility",
    "shop",
]);
```

- [ ] **Step 3: Verify TypeScript compiles**

```bash
cd /Users/ictsaigon.vn/mywork/letslive/web && npx tsc --noEmit 2>&1 | head -30
```

Expected: no errors.

- [ ] **Step 4: Verify gift flow in browser**

1. Log in and navigate to another user's profile page.
2. Confirm the 🎁 Gift button appears below the Follow button.
3. Click Gift button — modal opens showing shop items.
4. Click an item — toast appears confirming gift sent (or error if insufficient balance).
5. Navigate to `/en/users/<recipient-id>/gifts` — sent gift appears in collection.

- [ ] **Step 5: Commit**

```bash
git add "web/app/[lng]/(main)/users/[userId]/gift-modal.tsx" \
        "web/app/[lng]/(main)/users/[userId]/profile-header.tsx"
git commit -m "feat(web): add gift button and modal on user profile

Refs: docs/superpowers/specs/2026-06-24-shop-gift-design.md"
```

---

## Self-Review Checklist

- [x] **Spec coverage:** shop catalog ✓, user gifts page ✓, inventory page ✓, gift button on profile ✓, quick-send flow ✓, send-from-inventory flow (API wired, no dedicated UI button — spec says gift button on profile uses quick-send; inventory send is a follow-up)
- [x] **Placeholders:** none — all code blocks contain real implementations
- [x] **Type consistency:** `ShopItem`, `UserInventory`, `Gift`, `PurchaseRequest`, `PurchaseResponse`, `SendGiftRequest` defined in Task 1 and used by exact same names in Tasks 2–7
- [x] **Kong paths:** no `/v1/` prefix — Kong strips it (`/shop/items`, `/shop/purchase`, `/user/me/inventory`, `/user/{userId}/gifts/received`, `/gifts`)
- [x] **Auth:** inventory page auth-gated via wallet `layout.tsx` (checks `user`); shop + gifts pages are public; gift modal checks `me?.id !== user.id`
