import { FetchError } from "../../types/fetch-error";
import { fetchClient } from "../../utils/fetchClient";
import jwt from "jsonwebtoken";

export async function SignUp(body: {
  email: string;
  username: string;
  password: string;
  turnstileToken: string;
}): Promise<{ fetchError?: FetchError }> {
  try {
    await fetchClient<void>("/auth/signup", {
      method: "POST",
      body: JSON.stringify({
        email: body.email,
        username: body.username,
        password: body.password,
        turnstileToken: body.turnstileToken,
      }),
    });

    return {};
  } catch (error) {
    return { fetchError: error as FetchError };
  }
}

export async function LogIn(body: {
  email: string;
  password: string;
  turnstileToken: string;
}): Promise<{ fetchError?: FetchError }> {
  try {
    await fetchClient<void>("/auth/login", {
      method: "POST",
      body: JSON.stringify({
        email: body.email,
        password: body.password,
        turnstileToken: body.turnstileToken,
      }),
    });
    return {};
  } catch (error) {
    return { fetchError: error as FetchError };
  }
}

export async function Logout(): Promise<{ fetchError?: FetchError }> {
  try {
    await fetchClient<void>("/auth/logout", {
      method: "DELETE",
    });
    return {};
  } catch (error) {
    return { fetchError: error as FetchError };
  }
}

export async function ChangePassword(body: {
  oldPassword: string;
  newPassword: string;
}): Promise<{ fetchError?: FetchError }> {
  try {
    await fetchClient<void>("/auth/password", {
      method: "PATCH",
      body: JSON.stringify({
        oldPassword: body.oldPassword,
        newPassword: body.newPassword,
      }),
    });
    return {};
  } catch (error) {
    return { fetchError: error as FetchError };
  }
}
