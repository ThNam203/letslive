import { useMutation, useQueryClient } from "@tanstack/react-query";
import { RequestToGenerateNewAPIKey } from "@/lib/api/user";
import { unwrapResponse } from "@/lib/api/api-error";
import useUser from "@/hooks/user";
import { ME_PROFILE_QUERY_KEY } from "./use-users";
import { MeUser } from "@/types/user";

export function useGenerateApiKey() {
    const queryClient = useQueryClient();
    const updateUser = useUser((state) => state.updateUser);

    return useMutation({
        mutationFn: async () =>
            unwrapResponse(await RequestToGenerateNewAPIKey()),
        onSuccess: (streamAPIKey) => {
            updateUser({ streamAPIKey });
            queryClient.setQueryData<MeUser | null>(
                ME_PROFILE_QUERY_KEY,
                (prev) => (prev ? { ...prev, streamAPIKey } : prev),
            );
        },
    });
}
