import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Logout } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/api-error";
import useUser from "@/hooks/user";

/**
 * Ends the session.
 *
 * Clears the whole query cache on the way out: everything in it was fetched
 * as the user signing out, and the next person to sign in on this browser
 * must not be shown any of it.
 */
export function useLogout() {
    const queryClient = useQueryClient();
    const clearUser = useUser((state) => state.clearUser);

    return useMutation({
        mutationFn: async () => {
            // the endpoint answers 204 with no envelope of its own
            const res = await Logout();
            if (res.statusCode !== 204) throw new ApiError(res);
            return res;
        },
        onSuccess: () => {
            clearUser();
            queryClient.clear();
        },
    });
}
