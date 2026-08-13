import { useQuery } from "@tanstack/react-query";
import { GetMyInventory } from "@/lib/api/gift";
import { unwrapResponse } from "@/lib/api/api-error";

export const INVENTORY_QUERY_KEY = ["wallet", "inventory"] as const;

export function useMyInventory(enabled: boolean) {
    return useQuery({
        queryKey: INVENTORY_QUERY_KEY,
        queryFn: async () => unwrapResponse(await GetMyInventory()),
        enabled,
    });
}
