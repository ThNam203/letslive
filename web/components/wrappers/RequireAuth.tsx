"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import useUser from "@/hooks/user";

export default function RequireAuth({
    children,
    fallback = null,
}: {
    children: React.ReactNode;
    fallback?: React.ReactNode;
}) {
    const { user, isLoading } = useUser();
    const router = useRouter();
    const pathname = usePathname();

    useEffect(() => {
        if (isLoading || user) return;
        router.replace(`/login?redirectUrl=${encodeURIComponent(pathname)}`);
    }, [isLoading, user, pathname, router]);

    if (isLoading || !user) return <>{fallback}</>;

    return <>{children}</>;
}
