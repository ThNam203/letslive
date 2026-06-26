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
