import { ApiResponse } from "@/types/fetch-response";
import GLOBAL from "@/lib/global";

type WithStatusCode<T> = T & { statusCode: number };

export async function adminFetch<T>(
    path: string,
    options: RequestInit = {},
): Promise<WithStatusCode<ApiResponse<T>>> {
    const url = path.startsWith("http") ? path : GLOBAL.ADMIN_API_URL + path;

    const response = await fetch(url, {
        credentials: "include",
        ...options,
        headers: {
            "Content-Type": "application/json",
            ...(options.headers ?? {}),
        },
    });

    const data = (await response.json()) as ApiResponse<T>;
    return { ...data, statusCode: response.status };
}
