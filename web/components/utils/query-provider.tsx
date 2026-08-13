"use client";

import { useState } from "react";
import {
    QueryCache,
    QueryClient,
    QueryClientProvider,
} from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { ApiError } from "@/lib/api/api-error";
import { toast } from "@/components/utils/toast";
import i18next from "@/lib/i18n/i18next";

function handleQueryError(error: unknown) {
    if (error instanceof ApiError) {
        toast.error(i18next.t(`api-response:${error.key}`), {
            toastId: error.requestId,
        });
        return;
    }
    toast.error(i18next.t("fetch-error:client_fetch_error"));
}

export default function QueryProvider({
    children,
}: {
    children: React.ReactNode;
}) {
    const [queryClient] = useState(
        () =>
            new QueryClient({
                queryCache: new QueryCache({ onError: handleQueryError }),
                defaultOptions: {
                    queries: {
                        staleTime: 30_000,
                        retry: 1,
                    },
                },
            }),
    );

    return (
        <QueryClientProvider client={queryClient}>
            {children}
            {process.env.NODE_ENV === "development" && (
                <ReactQueryDevtools initialIsOpen={false} />
            )}
        </QueryClientProvider>
    );
}
