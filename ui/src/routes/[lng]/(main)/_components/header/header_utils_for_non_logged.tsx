"use client";

import LanguageSwitch from "@/src/components/utils/language-switch";
import ThemeSwitch from "@/src/components/utils/theme-switch";
import useUser from "@/src/hooks/user";

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
