"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useCallback, useLayoutEffect, useRef, useState } from "react";
import { cn } from "@/utils/cn";
import useT from "@/hooks/use-translation";
import IconChevronDown from "@/components/icons/chevron-down";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export type NavTab = { name: string; href: string };

const tabBaseClassName =
    "hover:text-primary relative inline-block py-4 text-center text-sm whitespace-nowrap transition-colors";
const moreBaseClassName =
    "hover:text-primary relative inline-flex items-center justify-center gap-1 py-4 text-sm whitespace-nowrap transition-colors outline-none";

function activeClassName(isActive: boolean) {
    return isActive
        ? "text-primary border-primary border-b-2"
        : "text-foreground";
}

// The only interactive part of a tabbed page: the active tab follows the URL.
// Tabs that don't fit the width are moved into a "More" dropdown, like GitHub.
export default function NavTabs({
    tabs,
    tabClassName,
}: {
    tabs: NavTab[];
    tabClassName?: string;
}) {
    const pathname = usePathname();
    const { t } = useT("common");
    const navRef = useRef<HTMLElement>(null);
    // Invisible copy of every tab (plus the More button) used to measure widths.
    const measureRef = useRef<HTMLUListElement>(null);
    const [visibleCount, setVisibleCount] = useState(tabs.length);

    const recompute = useCallback(() => {
        const nav = navRef.current;
        const measure = measureRef.current;
        if (!nav || !measure) return;

        const items = Array.from(measure.children) as HTMLElement[];
        const moreWidth = items.pop()!.getBoundingClientRect().width;
        const widths = items.map((el) => el.getBoundingClientRect().width);
        const available = nav.clientWidth;

        if (widths.reduce((sum, w) => sum + w, 0) <= available) {
            setVisibleCount(widths.length);
            return;
        }

        let used = moreWidth;
        let count = 0;
        while (count < widths.length && used + widths[count] <= available) {
            used += widths[count];
            count++;
        }
        setVisibleCount(count);
    }, []);

    useLayoutEffect(() => {
        recompute();
        const observer = new ResizeObserver(recompute);
        if (navRef.current) observer.observe(navRef.current);
        // Tab widths change when labels change (locale switch, font load).
        if (measureRef.current) observer.observe(measureRef.current);
        return () => observer.disconnect();
    }, [recompute]);

    const visibleTabs = tabs.slice(0, visibleCount);
    const overflowTabs = tabs.slice(visibleCount);
    const isOverflowActive = overflowTabs.some((tab) =>
        pathname.endsWith(tab.href),
    );

    return (
        <nav
            ref={navRef}
            className="border-border relative overflow-x-clip border-b"
        >
            <ul className="flex">
                {visibleTabs.map((tab) => {
                    const isActive = pathname.endsWith(tab.href);
                    return (
                        <li key={tab.href} className="shrink-0">
                            <Link
                                href={tab.href}
                                aria-current={isActive ? "page" : undefined}
                                className={cn(
                                    tabBaseClassName,
                                    tabClassName,
                                    activeClassName(isActive),
                                )}
                            >
                                {tab.name}
                            </Link>
                        </li>
                    );
                })}
                {overflowTabs.length > 0 && (
                    <li className="shrink-0">
                        <DropdownMenu modal={false}>
                            <DropdownMenuTrigger
                                className={cn(
                                    moreBaseClassName,
                                    tabClassName,
                                    activeClassName(isOverflowActive),
                                )}
                            >
                                {t("more")}
                                <IconChevronDown className="h-4 w-4" />
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="border-border"
                            >
                                {overflowTabs.map((tab) => {
                                    const isActive = pathname.endsWith(
                                        tab.href,
                                    );
                                    return (
                                        <DropdownMenuItem
                                            key={tab.href}
                                            asChild
                                        >
                                            <Link
                                                href={tab.href}
                                                aria-current={
                                                    isActive
                                                        ? "page"
                                                        : undefined
                                                }
                                                className={cn(
                                                    "cursor-pointer",
                                                    isActive && "text-primary",
                                                )}
                                            >
                                                {tab.name}
                                            </Link>
                                        </DropdownMenuItem>
                                    );
                                })}
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </li>
                )}
            </ul>

            <ul
                ref={measureRef}
                aria-hidden
                className="pointer-events-none invisible absolute top-0 left-0 flex w-max"
            >
                {tabs.map((tab) => (
                    <li key={tab.href} className="shrink-0">
                        <span className={cn(tabBaseClassName, tabClassName)}>
                            {tab.name}
                        </span>
                    </li>
                ))}
                <li className="shrink-0">
                    <span className={cn(moreBaseClassName, tabClassName)}>
                        {t("more")}
                        <IconChevronDown className="h-4 w-4" />
                    </span>
                </li>
            </ul>
        </nav>
    );
}
