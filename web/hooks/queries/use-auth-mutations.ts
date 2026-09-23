import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
    LogIn,
    Logout,
    RequestToSendVerification,
    SignUp,
} from "@/lib/api/auth";
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

/**
 * Signing in and signing up both fail in the same way — a bad credential, a
 * spent captcha — and the query client's error handler already reports that
 * with the server's own message. These mutations only throw; the forms decide
 * what to reset.
 */
export function useLogin() {
    return useMutation({
        mutationFn: async (credentials: {
            email: string;
            password: string;
            turnstileToken: string;
        }) => {
            const res = await LogIn(credentials);
            if (!res.success) throw new ApiError(res);
            return res;
        },
    });
}

export function useSignup() {
    return useMutation({
        mutationFn: async (registration: {
            email: string;
            username: string;
            password: string;
            turnstileToken: string;
            otpCode: string;
        }) => {
            const res = await SignUp(registration);
            if (!res.success) throw new ApiError(res);
            return res;
        },
    });
}

export function useRequestEmailVerification() {
    return useMutation({
        mutationFn: async (input: {
            email: string;
            turnstileToken: string;
        }) => {
            const res = await RequestToSendVerification(
                input.email,
                input.turnstileToken,
            );
            if (!res.success) throw new ApiError(res);
            return res;
        },
    });
}
