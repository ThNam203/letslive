import { useMutation, useQueryClient } from "@tanstack/react-query";
import { UpdateProfile } from "@/lib/api/user";
import { ApiError } from "@/lib/api/api-error";
import useUser from "@/hooks/user";
import { MeUser } from "@/types/user";
import { ME_PROFILE_QUERY_KEY } from "./use-users";

/**
 * Saves a partial change to the signed-in user's profile.
 *
 * Resolves with the whole response envelope rather than just the profile,
 * because call sites announce success with the server's own i18n key and
 * request id. Failures throw, so the query client's shared error handler
 * reports them.
 */
export function useUpdateProfile() {
    const queryClient = useQueryClient();
    const updateUser = useUser((state) => state.updateUser);

    return useMutation({
        mutationFn: async (update: Partial<MeUser>) => {
            const res = await UpdateProfile(update);
            if (!res.success) throw new ApiError(res);
            return res;
        },
        onSuccess: (res) => {
            if (!res.data) return;
            updateUser(res.data);
            queryClient.setQueryData(ME_PROFILE_QUERY_KEY, res.data);
        },
    });
}

/**
 * Refetches the signed-in user after the session changes (sign-in, sign-up).
 * The profile query is mounted app-wide, so invalidating it is enough to
 * refill both the cache and the user store.
 */
export function useRefreshMeProfile() {
    const queryClient = useQueryClient();
    return () =>
        queryClient.invalidateQueries({ queryKey: ME_PROFILE_QUERY_KEY });
}
