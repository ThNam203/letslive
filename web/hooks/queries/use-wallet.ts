import { useQuery } from "@tanstack/react-query";
import { GetMyWallet, GetTransactions } from "@/lib/api/wallet";
import { unwrapResponse } from "@/lib/api/api-error";

export const WALLET_BALANCE_QUERY_KEY = ["wallet", "balance"] as const;

export function useWalletBalance(enabled: boolean) {
    return useQuery({
        queryKey: WALLET_BALANCE_QUERY_KEY,
        queryFn: async () => unwrapResponse(await GetMyWallet()),
        enabled,
    });
}

export function useRecentTransactions(enabled: boolean, limit: number = 5) {
    return useQuery({
        queryKey: ["wallet", "transactions", "recent", limit],
        queryFn: async () => unwrapResponse(await GetTransactions(0, limit)),
        enabled,
    });
}
