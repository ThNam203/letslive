"use client";

import LanguageSwitch from "@/components/utils/language-switch";
import ThemeSwitch from "@/components/utils/theme-switch";
import useUser from "@/hooks/user";

export default function HeaderUtilsForNonLogged() {
    const user = useUser((state) => state.user);
    const isLoading = useUser((state) => state.isLoading);

    // hidden until the session is known, so signed-in users never see them flash
    if (user || isLoading) return null;
    return (
        <>
            <LanguageSwitch className="h-8" />
            <ThemeSwitch className="h-8" />
        </>
    );
}
