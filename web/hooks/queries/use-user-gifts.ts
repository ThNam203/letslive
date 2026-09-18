import { useQuery } from "@tanstack/react-query";
import { GetUserGiftsReceived } from "@/lib/api/gift";
import { unwrapResponse } from "@/lib/api/api-error";

export function userGiftsReceivedQueryKey(userId: string) {
    return ["gifts", "received", userId] as const;
}

export function useUserGiftsReceived(userId: string | undefined) {
    return useQuery({
        queryKey: userGiftsReceivedQueryKey(userId ?? ""),
        queryFn: async () =>
            unwrapResponse(await GetUserGiftsReceived(userId as string)),
        enabled: Boolean(userId),
    });
}
