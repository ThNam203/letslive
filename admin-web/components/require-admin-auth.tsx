"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { GetMe } from "@/lib/api/auth";
import { unwrapResponse } from "@/lib/api/api-error";

export default function RequireAdminAuth({
    children,
}: {
    children: React.ReactNode;
}) {
    const router = useRouter();
    const [redirected, setRedirected] = useState(false);

    const { data: me, isLoading, isError } = useQuery({
        queryKey: ["admin", "me"],
        queryFn: async () => unwrapResponse(await GetMe()),
        retry: false,
    });

    useEffect(() => {
        if (!isLoading && isError && !redirected) {
            setRedirected(true);
            router.replace("/login");
        }
    }, [isLoading, isError, redirected, router]);

    if (isLoading) {
        return <div className="p-8 text-muted">Loading...</div>;
    }

    if (isError || !me) {
        return null;
    }

    return <>{children}</>;
}
