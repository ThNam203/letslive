"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Trans } from "react-i18next";
import { Button } from "@/components/ui/button";
import useT from "@/hooks/use-translation";

const COOKIE_CONSENT_STORAGE_KEY = "cookie_consent_accepted";

const CookieConsentBanner = () => {
    const { t, i18n } = useT(["common", "legal"]);
    const [isVisible, setIsVisible] = useState(false);

    useEffect(() => {
        queueMicrotask(() => {
            const hasAccepted = localStorage.getItem(
                COOKIE_CONSENT_STORAGE_KEY,
            );
            if (!hasAccepted) setIsVisible(true);
        });
    }, []);

    const handleAccept = () => {
        localStorage.setItem(COOKIE_CONSENT_STORAGE_KEY, "true");
        setIsVisible(false);
    };

    if (!isVisible) return null;

    return (
        <div className="border-border bg-background text-foreground fixed inset-x-0 bottom-0 z-50 flex flex-col items-center gap-3 border-t p-4 shadow-lg sm:flex-row sm:justify-between">
            <p className="text-sm">
                <Trans
                    t={t}
                    i18nKey="legal:cookie_consent_message"
                    i18n={i18n}
                    components={{
                        policyLink: (
                            <Link
                                href="/cookie-policy"
                                className="underline underline-offset-2"
                            />
                        ),
                    }}
                />
            </p>
            <Button onClick={handleAccept} className="w-full sm:w-auto">
                {t("common:accept")}
            </Button>
        </div>
    );
};

export default CookieConsentBanner;
