import { adminFetch } from "@/lib/api/admin-fetch";
import { ApiResponse } from "@/types/fetch-response";
import { AdminMe } from "@/types/admin";

export async function Login(
    email: string,
    password: string,
): Promise<ApiResponse<null>> {
    return adminFetch<null>("/admin/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
    });
}

export async function GetMe(): Promise<ApiResponse<AdminMe>> {
    return adminFetch<AdminMe>("/admin/me");
}
