"use client";

import i18next from "@/lib/i18n/i18next";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { UseTranslationOptions } from "react-i18next";
import { useParams, usePathname } from "next/navigation";

const runsOnServerSide = typeof window === "undefined";

function useT(
    ns: string | string[] = "common",
    options?: UseTranslationOptions<undefined>,
) {
    const pathname = usePathname();
    const lng = useParams()?.lng ?? pathname.split("/")[1];
    if (typeof lng !== "string")
        throw new Error("useT is only available inside /app/[lng]");

    const [activeLng, setActiveLng] = useState(i18next.resolvedLanguage);

    useEffect(() => {
        if (runsOnServerSide) return;
        if (activeLng === i18next.resolvedLanguage) return;
        queueMicrotask(() => setActiveLng(i18next.resolvedLanguage));
    }, [activeLng, i18next.resolvedLanguage]);

    useEffect(() => {
        if (runsOnServerSide) return;
        if (!lng || i18next.resolvedLanguage === lng) return;
        i18next.changeLanguage(lng);
    }, [lng]);

    if (runsOnServerSide && i18next.resolvedLanguage !== lng) {
        i18next.changeLanguage(lng);
    }

    return useTranslation(ns, options);
}

export default useT;
