"use client";

import { useQuery } from "@tanstack/react-query";
import RequireAdminAuth from "@/components/require-admin-auth";
import { GetMe } from "@/lib/api/auth";
import { unwrapResponse } from "@/lib/api/api-error";

function HomeContent() {
    const { data: me } = useQuery({
        queryKey: ["admin", "me"],
        queryFn: async () => unwrapResponse(await GetMe()),
    });

    return (
        <div className="p-8">
            <h1 className="text-xl font-semibold">letslive admin</h1>
            <p className="text-muted mt-2">Logged in as {me?.email}.</p>
        </div>
    );
}

export default function HomePage() {
    return (
        <RequireAdminAuth>
            <HomeContent />
        </RequireAdminAuth>
    );
}
