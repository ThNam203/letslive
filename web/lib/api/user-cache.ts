import { ApiResponse } from "@/types/fetch-response";
import { PublicUser } from "@/types/user";
import { GetUserById } from "./user";

type CachedResponse = ApiResponse<PublicUser> & { statusCode: number };

const CACHE_TTL_MS = 60_000;

const inflight = new Map<string, Promise<CachedResponse>>();
const cache = new Map<string, { value: CachedResponse; expiresAt: number }>();

export function GetUserByIdCached(
    userId: string,
): Promise<CachedResponse> {
    const now = Date.now();
    const entry = cache.get(userId);
    if (entry && entry.expiresAt > now) {
        return Promise.resolve(entry.value);
    }

    const existing = inflight.get(userId);
    if (existing) return existing;

    const promise = GetUserById(userId)
        .then((res) => {
            if (res.success) {
                cache.set(userId, {
                    value: res,
                    expiresAt: Date.now() + CACHE_TTL_MS,
                });
            }
            return res;
        })
        .finally(() => {
            inflight.delete(userId);
        });

    inflight.set(userId, promise);
    return promise;
}
