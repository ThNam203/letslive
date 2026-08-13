import { useQuery } from "@tanstack/react-query";
import { GetShopItems } from "@/lib/api/shop";
import { unwrapResponse } from "@/lib/api/api-error";

export function useShopItems(options?: { enabled?: boolean }) {
    return useQuery({
        queryKey: ["shop-items"],
        queryFn: async () => unwrapResponse(await GetShopItems()),
        enabled: options?.enabled,
    });
}
