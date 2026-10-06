"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { GetMe } from "@/lib/api/auth";
import { ApiError, unwrapResponse } from "@/lib/api/api-error";

function getErrorMessage(error: unknown): string {
    if (error instanceof Error) {
        return error.message;
    }
    return "Something went wrong.";
}

export default function RequireAdminAuth({
    children,
}: {
    children: React.ReactNode;
}) {
    const router = useRouter();

    const {
        data: me,
        isLoading,
        isError,
        error,
        refetch,
    } = useQuery({
        queryKey: ["admin", "me"],
        queryFn: async () => unwrapResponse(await GetMe()),
        retry: false,
    });

    // Only a genuine 401 means "not logged in" — any other failure (500, a
    // Kong 502, a network error) must not be treated as a logged-out state,
    // or an already-logged-in admin gets bounced during an outage.
    const isUnauthenticated =
        isError && error instanceof ApiError && error.statusCode === 401;

    useEffect(() => {
        if (isUnauthenticated) {
            router.replace("/login");
        }
    }, [isUnauthenticated, router]);

    if (isLoading) {
        return <div className="p-8 text-muted">Loading...</div>;
    }

    if (isUnauthenticated) {
        return null;
    }

    if (isError) {
        return (
            <div className="p-8">
                <p className="text-destructive text-sm">
                    {getErrorMessage(error)}
                </p>
                <button
                    type="button"
                    onClick={() => refetch()}
                    className="border-border mt-4 rounded border px-3 py-2 text-sm"
                >
                    Retry
                </button>
            </div>
        );
    }

    if (!me) {
        return null;
    }

    return <>{children}</>;
}
