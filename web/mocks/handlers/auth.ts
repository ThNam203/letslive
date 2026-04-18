import { http } from "msw";
import { API_BASE, ok, noContent, badRequest } from "../utils";
import { meUser } from "../db";

export const authHandlers = [
    // POST /auth/login — always succeeds, sets a fake cookie via header
    http.post(`${API_BASE}/auth/login`, async ({ request }) => {
        const body = (await request.json()) as any;
        if (!body?.email || !body?.password) {
            return badRequest(
                "res_err_invalid_input",
                "Email and password are required",
            );
        }
        const response = ok<void>(undefined);
        // Set fake auth cookies so the app treats the user as logged in
        response.headers.append(
            "Set-Cookie",
            "ACCESS_TOKEN=mock-access-token; Path=/; SameSite=Lax",
        );
        response.headers.append(
            "Set-Cookie",
            "REFRESH_TOKEN=mock-refresh-token; Path=/; SameSite=Lax",
        );
        return response;
    }),

    // POST /auth/signup — always succeeds
    http.post(`${API_BASE}/auth/signup`, async () => {
        return ok<void>(undefined);
    }),

    // DELETE /auth/logout
    http.delete(`${API_BASE}/auth/logout`, async () => {
        const response = noContent();
        // Clear fake cookies
        response.headers.append(
            "Set-Cookie",
            "ACCESS_TOKEN=; Path=/; Max-Age=0",
        );
        response.headers.append(
            "Set-Cookie",
            "REFRESH_TOKEN=; Path=/; Max-Age=0",
        );
        return response;
    }),

    // PATCH /auth/password
    http.patch(`${API_BASE}/auth/password`, async () => {
        return ok<void>(undefined);
    }),

    // POST /auth/verify-email
    http.post(`${API_BASE}/auth/verify-email`, async () => {
        return ok<void>(undefined);
    }),

    // POST /auth/refresh-token — always succeeds in mock mode
    http.post(`${API_BASE}/auth/refresh-token`, async () => {
        return ok<void>(undefined);
    }),
];
