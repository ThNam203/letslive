"use client";

import i18next from "@/src/lib/i18n/i18next";
import { useEffect, useState } from "react";
import { useTranslation } from "next-i18next";
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
    if (runsOnServerSide && i18next.resolvedLanguage !== lng) {
        i18next.changeLanguage(lng);
    } else {
        const [activeLng, setActiveLng] = useState(i18next.resolvedLanguage);
        useEffect(() => {
            if (activeLng === i18next.resolvedLanguage) return;
            setActiveLng(i18next.resolvedLanguage);
        }, [activeLng, i18next.resolvedLanguage]);
        useEffect(() => {
            if (!lng || i18next.resolvedLanguage === lng) return;
            i18next.changeLanguage(lng);
        }, [lng, i18next]);
    }
    return useTranslation(ns, options);
}

export default useT;
