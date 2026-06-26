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
