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
