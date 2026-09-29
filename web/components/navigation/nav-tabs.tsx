"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/utils/cn";

export type NavTab = { name: string; href: string };

// The only interactive part of a tabbed page: the active tab follows the URL.
export default function NavTabs({
    tabs,
    tabClassName,
}: {
    tabs: NavTab[];
    tabClassName?: string;
}) {
    const pathname = usePathname();

    return (
        <nav className="border-border border-b">
            <ul className="flex">
                {tabs.map((tab) => {
                    const isActive = pathname.endsWith(tab.href);
                    return (
                        <li key={tab.href}>
                            <Link
                                href={tab.href}
                                aria-current={isActive ? "page" : undefined}
                                className={cn(
                                    "hover:text-primary relative inline-block py-4 text-center text-sm transition-colors",
                                    tabClassName,
                                    isActive
                                        ? "text-primary border-primary border-b-2"
                                        : "text-foreground",
                                )}
                            >
                                {tab.name}
                            </Link>
                        </li>
                    );
                })}
            </ul>
        </nav>
    );
}
