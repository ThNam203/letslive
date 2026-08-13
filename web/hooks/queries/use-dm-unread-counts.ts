import { useQuery } from "@tanstack/react-query";
import { GetUnreadCounts } from "@/lib/api/dm";
import { unwrapResponse } from "@/lib/api/api-error";

export const DM_UNREAD_COUNTS_QUERY_KEY = ["dm", "unread-counts"] as const;

export function useDmUnreadCounts(enabled: boolean) {
    return useQuery({
        queryKey: DM_UNREAD_COUNTS_QUERY_KEY,
        queryFn: async () => unwrapResponse(await GetUnreadCounts()),
        enabled,
        refetchInterval: 30_000,
        refetchIntervalInBackground: false,
    });
}
