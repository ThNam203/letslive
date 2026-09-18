import { useQuery } from "@tanstack/react-query";
import { GetMeProfile, GetUserById } from "@/lib/api/user";
import { ApiError, unwrapResponse } from "@/lib/api/api-error";
import useUser from "@/hooks/user";
import { MeUser } from "@/types/user";

export function publicUserQueryKey(userId: string) {
    return ["users", "public", userId] as const;
}

export const ME_PROFILE_QUERY_KEY = ["users", "me"] as const;

// Public profiles change rarely and the same user is requested from several
// cards on one screen, so a longer stale window keeps those renders off the
// network without going stale in a way anyone notices.
const PUBLIC_USER_STALE_TIME_MS = 60_000;

export function usePublicUser(userId: string | undefined, enabled = true) {
    return useQuery({
        queryKey: publicUserQueryKey(userId ?? ""),
        queryFn: async () =>
            unwrapResponse(await GetUserById(userId as string)),
        enabled: Boolean(userId) && enabled,
        staleTime: PUBLIC_USER_STALE_TIME_MS,
    });
}

/**
 * The signed-in user's own profile.
 *
 * Writes into the zustand user store as it resolves: the store is what the
 * rest of the app reads, and this query is the only thing that fills it.
 * A 401 resolves to `null` rather than throwing, because "nobody is signed
 * in" is an ordinary state of the app, not a failure worth a toast.
 */
export function useMeProfile() {
    const setUser = useUser((state) => state.setUser);
    const setIsLoading = useUser((state) => state.setIsLoading);

    return useQuery({
        queryKey: ME_PROFILE_QUERY_KEY,
        queryFn: async (): Promise<MeUser | null> => {
            // The store's loading flag gates the auth redirects in
            // RequireAuth and the settings/wallet layouts, so it is kept in
            // step with the request itself rather than with a render pass.
            setIsLoading(true);
            try {
                const res = await GetMeProfile();
                if (!res.success) {
                    if (res.statusCode === 401) {
                        setUser(null);
                        return null;
                    }
                    throw new ApiError(res);
                }
                const profile = res.data ?? null;
                setUser(profile);
                return profile;
            } finally {
                setIsLoading(false);
            }
        },
        retry: false,
    });
}
