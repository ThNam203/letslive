import { FetchError } from "@/types/fetch-error";
import { User } from "@/types/user";
import { fetchClient } from "@/utils/fetchClient";

export async function SignUp(body: {
    email: string;
    username: string;
    password: string;
}): Promise<{ fetchError?: FetchError }> {
    try {
        await fetchClient<User>("/auth/signup", {
            method: "POST",
            body: JSON.stringify({
                email: body.email,
                username: body.username,
                password: body.password,
            }),
        });
        return {};
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function LogIn(body: {
    email: string,
    password: string
}): Promise<{ fetchError?: FetchError }> {
    try {
        await fetchClient<void>("/auth/login", {
            method: "POST",
            body: JSON.stringify({
                email: body.email,
                password: body.password,
            }),
        });
        return {};
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}
