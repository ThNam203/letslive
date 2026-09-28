"use client";

import useUser from "@/hooks/user";
import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import i18next from "@/lib/i18n/i18next";
import { isSupportedLocale } from "@/lib/i18n/settings";
import { switchLocale } from "@/lib/i18n/switch-locale";
import { useMeProfile } from "@/hooks/queries/use-users";

export default function UserInformationWrapper({
    children,
}: {
    children: React.ReactNode;
}) {
    const user = useUser((state) => state.user);
    const router = useRouter();
    const pathname = usePathname();

    // Fetches the signed-in user once per session and fills the user store;
    // everything below reads the store, not this query.
    const { data: profile } = useMeProfile();

    // hydrate FE locale from the user's saved preference (login or session-restore)
    const savedLocale = user?.locale;
    useEffect(() => {
        if (!isSupportedLocale(savedLocale)) return;
        if (i18next.language === savedLocale) return;
        // no syncLocale: this value came from the DB in the first place
        switchLocale(router, savedLocale);
    }, [savedLocale, router]);

    // A profile with no username has never finished sign-up
    useEffect(() => {
        if (!profile || profile.username !== "") return;
        if (pathname.includes("account-setup")) return;
        router.push("/account-setup");
    }, [profile, pathname, router]);

    // Render children immediately - user fetch happens in background
    return <>{children}</>;
}
