import { ApiResponse } from "@/types/fetch-response";
import { fetchClient } from "@/utils/fetchClient";

export async function SignUp(body: {
    email: string;
    username: string;
    password: string;
    turnstileToken: string;
    otpCode: string;
}): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>("/auth/signup", {
        method: "POST",
        body: JSON.stringify({
            email: body.email,
            username: body.username,
            password: body.password,
            turnstileToken: body.turnstileToken,
            otpCode: body.otpCode,
        }),
    });
}

export async function LogIn(body: {
    email: string;
    password: string;
    turnstileToken: string;
}): Promise<ApiResponse<{ reactivationToken?: string }>> {
    return fetchClient<ApiResponse<{ reactivationToken?: string }>>(
        "/auth/login",
        {
            method: "POST",
            body: JSON.stringify({
                email: body.email,
                password: body.password,
                turnstileToken: body.turnstileToken,
            }),
        },
    );
}

export async function Logout(): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>("/auth/logout", {
        method: "DELETE",
    });
}

export async function LogoutAll(): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>("/auth/logout-all", {
        method: "DELETE",
    });
}

export async function ChangePassword(body: {
    oldPassword: string;
    newPassword: string;
}): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>("/auth/password", {
        method: "PATCH",
        body: JSON.stringify({
            oldPassword: body.oldPassword,
            newPassword: body.newPassword,
        }),
    });
}

export async function RequestToSendVerification(
    email: string,
    turnstileToken: string,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>("/auth/verify-email", {
        method: "POST",
        body: JSON.stringify({ email, turnstileToken }),
    });
}

export async function Reactivate(body: {
    reactivationToken: string;
}): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>("/auth/reactivate", {
        method: "POST",
        body: JSON.stringify({
            reactivationToken: body.reactivationToken,
        }),
    });
}
