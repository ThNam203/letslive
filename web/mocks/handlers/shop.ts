import { http } from "msw";
import { API_BASE, ok, notFound, badRequest } from "../utils";
import {
    shopItems,
    inventory,
    gifts,
    walletBalances,
    currencies,
    transactions,
    otherUsers,
    walletAccount,
    ME_USER_ID,
    uid,
    now,
} from "../db";
import {
    Gift,
    PurchaseRequest,
    PurchaseResponse,
    ShopItem,
    UserInventory,
} from "@/types/shop";
import {
    Transaction,
    TransactionStatus,
    TransactionType,
} from "@/types/wallet";

const DEFAULT_PAGE_SIZE = 20;

function pageParams(request: Request) {
    const url = new URL(request.url);
    const page = Math.max(
        0,
        parseInt(url.searchParams.get("page") ?? "0") || 0,
    );
    const limit =
        parseInt(url.searchParams.get("limit") ?? "") || DEFAULT_PAGE_SIZE;
    return { page, limit };
}

function paginate<T>(items: T[], page: number, limit: number) {
    return ok<T[]>(items.slice(page * limit, page * limit + limit), {
        page,
        page_size: limit,
        total: items.length,
    });
}

/**
 * Debits the mock wallet the way the finance service does: the price is in
 * whole units, balances are decimal strings at the currency's precision.
 * Returns false when the balance does not cover the cost.
 */
function debitWallet(item: ShopItem, quantity: number): boolean {
    const balance = walletBalances.find(
        (b) => b.currencyCode === item.currencyCode,
    );
    const precision =
        currencies.find((c) => c.code === item.currencyCode)?.precision ?? 2;
    const cost = item.price * quantity;
    const available = Number(balance?.balance ?? "0");
    if (!balance || available < cost) return false;

    const entryId = `entry-${uid()}`;
    balance.balance = (available - cost).toFixed(precision);
    balance.lastEntryId = entryId;

    const txId = `tx-${uid()}`;
    const tx: Transaction = {
        id: txId,
        type: TransactionType.PURCHASE,
        status: TransactionStatus.COMPLETED,
        reference: `purchase-${txId}`,
        actorId: ME_USER_ID,
        metadata: { shopItemId: item.id, quantity },
        createdAt: now(),
        entries: [
            {
                id: entryId,
                transactionId: txId,
                accountId: walletAccount.id,
                currencyCode: balance.currencyCode,
                amount: (-cost).toFixed(precision),
                createdAt: now(),
            },
        ],
    };
    transactions.push(tx);
    return true;
}

function addToInventory(shopItemId: string, quantity: number) {
    const existing = inventory.find((i) => i.shopItemId === shopItemId);
    if (existing) {
        existing.quantity += quantity;
        existing.updatedAt = now();
        return;
    }
    inventory.push({
        id: `inv-${uid()}`,
        userId: ME_USER_ID,
        shopItemId,
        quantity,
        updatedAt: now(),
    });
}

export const shopHandlers = [
    // GET /shop/items
    http.get(`${API_BASE}/shop/items`, () => ok<ShopItem[]>(shopItems)),

    // GET /shop/items/:id
    http.get(`${API_BASE}/shop/items/:id`, ({ params }) => {
        const item = shopItems.find((i) => i.id === params.id);
        if (!item)
            return notFound(
                "res_err_shop_item_not_found",
                "Shop item not found or unavailable.",
            );
        return ok<ShopItem>(item);
    }),

    // POST /shop/purchase: buys for yourself (inventory) or as a gift
    http.post(`${API_BASE}/shop/purchase`, async ({ request }) => {
        const body = (await request.json()) as Partial<PurchaseRequest>;
        const quantity = body.quantity ?? 0;
        if (!body.shopItemId || !Number.isInteger(quantity) || quantity < 1) {
            return badRequest("res_err_invalid_amount", "Invalid amount.");
        }

        const item = shopItems.find((i) => i.id === body.shopItemId);
        if (!item)
            return notFound(
                "res_err_shop_item_not_found",
                "Shop item not found or unavailable.",
            );

        const recipientId = body.recipientUserId;
        if (recipientId && !otherUsers.some((u) => u.id === recipientId)) {
            return notFound("res_err_user_not_found", "User not found.");
        }

        if (!debitWallet(item, quantity)) {
            return badRequest(
                "res_err_insufficient_balance",
                "Insufficient balance.",
            );
        }

        if (recipientId) {
            const gift: Gift = {
                id: `gift-${uid()}`,
                senderUserId: ME_USER_ID,
                recipientUserId: recipientId,
                shopItemId: item.id,
                quantity,
                message: body.message ?? null,
                sentAt: now(),
            };
            gifts.push(gift);
            return ok<PurchaseResponse>({
                giftId: gift.id,
                animationUrl: item.animationUrl,
            });
        }

        addToInventory(item.id, quantity);
        return ok<PurchaseResponse>({
            giftId: null,
            animationUrl: item.animationUrl,
        });
    }),

    // GET /user/me/inventory
    http.get(`${API_BASE}/user/me/inventory`, ({ request }) => {
        const { page, limit } = pageParams(request);
        const mine = inventory
            .filter((i) => i.userId === ME_USER_ID && i.quantity > 0)
            .sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));
        return paginate<UserInventory>(mine, page, limit);
    }),

    // GET /user/:userId/gifts/received
    http.get(
        `${API_BASE}/user/:userId/gifts/received`,
        ({ params, request }) => {
            const { page, limit } = pageParams(request);
            const received = gifts
                .filter((g) => g.recipientUserId === params.userId)
                .sort((a, b) => b.sentAt.localeCompare(a.sentAt));
            return paginate<Gift>(received, page, limit);
        },
    ),
];
