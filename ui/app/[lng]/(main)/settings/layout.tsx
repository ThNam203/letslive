"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { cn } from "@/utils/cn";
import useUser from "@/hooks/user";
import { useEffect, useState } from "react";
import IconLoader from "@/components/icons/loader";

const navItems = [
    { name: "Profile", href: "/settings/profile" },
    { name: "Security", href: "/settings/security" },
    { name: "Stream", href: "/settings/stream" },
    { name: "VODs", href: "/settings/vods" },
];

export default function SettingsNav({
    children,
}: Readonly<{ children: React.ReactNode }>) {
    const [isGettingUser, setIsGettingUser] = useState(true);
    const pathname = usePathname();
    const fetchUser = useUser((state) => state.fetchUser);
    const router = useRouter();

    useEffect(() => {
        setIsGettingUser(true);
        fetchUser()
            .catch(() => router.push("/login")) // TODO: should not redirect but show error
            .finally(() => {
                setIsGettingUser(false);
            });
    }, [fetchUser, router]);

    return (
        <div className="flex h-full flex-col bg-background text-foreground">
            <div className="max-w-7xl px-6">
                <div className="flex mt-6 items-center">
                    <h1 className="text-4xl font-bold">Settings</h1>
                    {isGettingUser && <IconLoader width="40" height="40"/>}
                </div>
                <nav className="border-b border-border">
                    <ul className="flex">
                        {navItems.map((item) => (
                            <li key={item.href}>
                                <Link
                                    href={item.href}
                                    className={cn(
                                        "relative inline-block w-20 py-4 text-center text-sm transition-colors hover:text-primary",
                                        pathname === item.href
                                            ? "text-primary"
                                            : "text-foreground",
                                    )}
                                >
                                    {item.name}
                                    {pathname === item.href && (
                                        <span className="absolute bottom-0 left-0 h-0.5 w-full bg-primary" />
                                    )}
                                </Link>
                            </li>
                        ))}
                    </ul>
                </nav>
            </div>
            <div className="flex-1 overflow-y-auto p-6 text-foreground">
                <div className="max-w-4xl space-y-8">
                    {!isGettingUser && children}
                </div>
            </div>
        </div>
    );
}
