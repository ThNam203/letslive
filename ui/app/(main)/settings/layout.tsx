"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { cn } from "../../../utils/cn";
import useUser from "@/hooks/user";
import { useEffect } from "react";

const navItems = [
    { name: "Profile", href: "/settings/profile" },
    { name: "Security", href: "/settings/security" },
    { name: "Stream", href: "/settings/stream" },
    { name: "VODs", href: "/settings/vods" }
];

export default function SettingsNav({
    children,
}: Readonly<{ children: React.ReactNode }>) {
    const pathname = usePathname();
    const user = useUser((state) => state.user)
    const fetchUser = useUser((state) => state.fetchUser)
    const router = useRouter();

    useEffect(() => {
        fetchUser().catch(() => router.push("/login"))
    }, [fetchUser, router])

    if (!user) return <p>Unauthenticated</p>

    return (
            <div className="overflow-y-auto h-full bg-background text-foreground">
                <div className="max-w-7xl px-6">
                    <h1 className="text-4xl font-bold py-6">Settings</h1>
                    <nav className="border-b border-border">
                        <ul className="flex flex-wrap gap-8">
                            {navItems.map((item) => (
                                <li key={item.href}>
                                    <Link
                                        href={item.href}
                                        className={cn(
                                            "inline-block py-4 text-sm relative hover:text-primary transition-colors",
                                            pathname === item.href
                                                ? "text-primary"
                                                : "text-foreground"
                                        )}
                                    >
                                        {item.name}
                                        {pathname === item.href && (
                                            <span className="absolute bottom-0 left-0 w-full h-0.5 bg-primary" />
                                        )}
                                    </Link>
                                </li>
                            ))}
                        </ul>
                    </nav>
                </div>
                {children}
            </div>
    );
}
