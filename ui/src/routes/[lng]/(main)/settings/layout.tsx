"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { cn } from "@/src/utils/cn";
import useUser from "@/src/hooks/user";
import IconLoader from "@/src/components/icons/loader";
import useT from "@/src/hooks/use-translation";

const getNavItems = (t: any) => [
    { name: t("settings:navigation.profile"), href: "/settings/profile" },
    { name: t("settings:navigation.security"), href: "/settings/security" },
    { name: t("settings:navigation.stream"), href: "/settings/stream" },
    { name: t("settings:navigation.vods"), href: "/settings/vods" },
];

export default function SettingsNav({
    children,
}: Readonly<{ children: React.ReactNode }>) {
    const pathname = usePathname();
    const { user, isLoading } = useUser();
    const { t } = useT(["settings", "fetch-error"]);
    const navItems = getNavItems(t);

    return (
        <div className="flex h-full flex-col bg-background text-foreground">
            <div className="max-w-7xl px-6">
                <div className="mt-6 flex items-center">
                    <h1 className="text-4xl font-bold">
                        {t("settings:page_title")}
                    </h1>
                    {isLoading && <IconLoader width="40" height="40" />}
                </div>
                <nav className="border-b border-border">
                    <ul className="flex">
                        {navItems.map((item) => {
                            const isActive = pathname.endsWith(item.href);
                            return (
                                <li key={item.href}>
                                    <Link
                                        href={item.href}
                                        className={cn(
                                            "relative inline-block w-20 py-4 text-center text-sm transition-colors hover:text-primary",
                                            isActive
                                                ? "text-primary border-b-2 border-primary"
                                                : "text-foreground",
                                        )}
                                    >
                                        {item.name}
                                    </Link>
                                </li>
                            );
                        })}
                    </ul>
                </nav>
            </div>
            <div className="flex-1 overflow-y-auto p-6 text-foreground">
                <div className="max-w-4xl space-y-8">
                    {isLoading ? null : !user ? (
                        <p>{t("settings:need_to_login")}</p>
                    ) : (
                        children
                    )}
                </div>
            </div>
        </div>
    );
}
