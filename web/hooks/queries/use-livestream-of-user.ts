import { useQuery } from "@tanstack/react-query";
import { GetLivestreamOfUser } from "@/lib/api/livestream";
import { unwrapResponse } from "@/lib/api/api-error";

export function livestreamOfUserQueryKey(userId: string) {
    return ["livestreams", "of-user", userId] as const;
}

// Resolves to null when the user is not live: the endpoint reports that as a
// success with no payload, not as an error.
export function useLivestreamOfUser(userId: string | undefined) {
    return useQuery({
        queryKey: livestreamOfUserQueryKey(userId ?? ""),
        queryFn: async () =>
            unwrapResponse(await GetLivestreamOfUser(userId as string)),
        enabled: Boolean(userId),
    });
}
