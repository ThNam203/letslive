"use client";

import { useEffect, useMemo } from "react";
import { I18nextProvider } from "react-i18next";
import { getI18nForLocale } from "@/lib/i18n/i18next";

export default function TranslationsProvider({
    lng,
    children,
}: {
    lng: string;
    children: React.ReactNode;
}) {
    const i18n = useMemo(() => getI18nForLocale(lng), [lng]);

    // the server may resolve a different locale than the browser holds, e.g.
    // after the cookie changed in another tab
    useEffect(() => {
        if (i18n.language !== lng) i18n.changeLanguage(lng);
    }, [i18n, lng]);

    return <I18nextProvider i18n={i18n}>{children}</I18nextProvider>;
}
