"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Login } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/api-error";

export default function LoginPage() {
    const router = useRouter();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState<string | null>(null);
    const [submitting, setSubmitting] = useState(false);

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault();
        setError(null);
        setSubmitting(true);

        try {
            const res = await Login(email, password);
            if (!res.success) {
                throw new ApiError(res);
            }
            router.replace("/");
        } catch (err) {
            setError(err instanceof ApiError ? err.message : "Login failed.");
        } finally {
            setSubmitting(false);
        }
    }

    return (
        <div className="flex min-h-screen items-center justify-center">
            <form
                onSubmit={handleSubmit}
                className="border-border w-full max-w-sm space-y-4 rounded-lg border p-6"
            >
                <h1 className="text-xl font-semibold">Admin login</h1>

                <div className="space-y-1">
                    <label className="text-muted text-sm" htmlFor="email">
                        Email
                    </label>
                    <input
                        id="email"
                        type="email"
                        required
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        className="border-border w-full rounded border bg-transparent px-3 py-2"
                    />
                </div>

                <div className="space-y-1">
                    <label className="text-muted text-sm" htmlFor="password">
                        Password
                    </label>
                    <input
                        id="password"
                        type="password"
                        required
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="border-border w-full rounded border bg-transparent px-3 py-2"
                    />
                </div>

                {error && <p className="text-destructive text-sm">{error}</p>}

                <button
                    type="submit"
                    disabled={submitting}
                    className="bg-primary text-primary-foreground w-full rounded px-3 py-2 disabled:opacity-50"
                >
                    {submitting ? "Logging in..." : "Log in"}
                </button>
            </form>
        </div>
    );
}
