import { useInfiniteQuery } from "@tanstack/react-query";
import { GetTransactions } from "@/lib/api/wallet";
import { nextPageParam, unwrapPage } from "@/lib/query/paginated";

const TRANSACTIONS_PAGE_SIZE = 20;

export function useTransactionsInfinite(enabled: boolean) {
    return useInfiniteQuery({
        queryKey: ["wallet", "transactions"],
        queryFn: async ({ pageParam }) =>
            unwrapPage(await GetTransactions(pageParam, TRANSACTIONS_PAGE_SIZE)),
        initialPageParam: 0,
        getNextPageParam: (_lastPage, allPages) =>
            nextPageParam(allPages, TRANSACTIONS_PAGE_SIZE),
        enabled,
    });
}
