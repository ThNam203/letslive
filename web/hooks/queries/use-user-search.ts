import { useQuery } from "@tanstack/react-query";
import { SearchUsersByUsername } from "@/lib/api/user";
import { unwrapResponse } from "@/lib/api/api-error";

// Typing back to a previous term is common, so results are kept briefly
// rather than re-requested on every keystroke that retraces one.
const SEARCH_STALE_TIME_MS = 30_000;

export function useUserSearch(query: string) {
    const trimmed = query.trim();

    return useQuery({
        queryKey: ["users", "search", trimmed] as const,
        queryFn: async () =>
            unwrapResponse(await SearchUsersByUsername(trimmed)),
        enabled: trimmed.length > 0,
        staleTime: SEARCH_STALE_TIME_MS,
    });
}
