"use client";

import LanguageSwitch from "@/components/utils/language-switch";
import ThemeSwitch from "@/components/utils/theme-switch";
import useUser from "@/hooks/user";

export default function HeaderUtilsForNonLogged() {
    const user = useUser((state) => state.user);

    if (user) return null;
    return (
        <>
            <LanguageSwitch className="h-8" />
            <ThemeSwitch className="h-8" />
        </>
    );
}
