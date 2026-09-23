import { useQuery } from "@tanstack/react-query";
import {
    GetAllVODsAsAuthor,
    GetPublicVODsOfUser,
    GetVODInformation,
} from "@/lib/api/vod";
import { unwrapResponse } from "@/lib/api/api-error";

export const AUTHOR_VODS_QUERY_KEY = ["vods", "author"] as const;

export function publicVodsOfUserQueryKey(userId: string) {
    return ["vods", "of-user", userId] as const;
}

export function vodQueryKey(vodId: string) {
    return ["vods", "detail", vodId] as const;
}

export function usePublicVodsOfUser(userId: string | undefined) {
    return useQuery({
        queryKey: publicVodsOfUserQueryKey(userId ?? ""),
        queryFn: async () =>
            unwrapResponse(await GetPublicVODsOfUser(userId as string)),
        enabled: Boolean(userId),
    });
}

// The author's own VODs, including private ones, so this is only meaningful
// once a user is signed in.
export function useAuthorVods(enabled: boolean) {
    return useQuery({
        queryKey: AUTHOR_VODS_QUERY_KEY,
        queryFn: async () => unwrapResponse(await GetAllVODsAsAuthor()),
        enabled,
    });
}

export function useVod(vodId: string | undefined) {
    return useQuery({
        queryKey: vodQueryKey(vodId ?? ""),
        queryFn: async () =>
            unwrapResponse(await GetVODInformation(vodId as string)),
        enabled: Boolean(vodId),
    });
}
