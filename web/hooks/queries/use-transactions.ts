import { useInfiniteQuery } from "@tanstack/react-query";
import { GetTransactions } from "@/lib/api/wallet";
import { unwrapResponse } from "@/lib/api/api-error";

const TRANSACTIONS_PAGE_SIZE = 20;

export function useTransactionsInfinite(enabled: boolean) {
    return useInfiniteQuery({
        queryKey: ["wallet", "transactions"],
        queryFn: async ({ pageParam }) =>
            unwrapResponse(
                await GetTransactions(pageParam, TRANSACTIONS_PAGE_SIZE),
            ),
        initialPageParam: 0,
        getNextPageParam: (lastPage, allPages) =>
            lastPage.length > 0 ? allPages.length : undefined,
        enabled,
    });
}
