"use client";

import { cn } from "@/utils/cn";
import useT from "@/hooks/use-translation";
import {
    I18N_LANGUAGE_COUNTRY_MAP,
    I18N_LANGUAGES,
} from "@/lib/i18n/settings";
import { switchLocale } from "@/lib/i18n/switch-locale";
import { usePathname, useRouter } from "next/navigation";

const LanguageList = ({ className }: { className?: string }) => {
    const { i18n } = useT();
    const router = useRouter();
    const pathname = usePathname();

    const handleChange = async (option: string) => {
        if (i18n.resolvedLanguage === option) return;
        await switchLocale(router, pathname, option);
    };

    return (
        <div className={cn("flex gap-2", className)}>
            {Object.values(I18N_LANGUAGES).map((lng) => (
                <button
                    key={lng}
                    onClick={() => handleChange(lng)}
                    className={cn(
                        "rounded-md border px-4 py-2 text-sm capitalize transition-all",
                        "border-border",
                        i18n.resolvedLanguage === lng
                            ? "ring-border font-bold ring-2"
                            : "",
                    )}
                >
                    <span>{I18N_LANGUAGE_COUNTRY_MAP[lng]}</span>
                </button>
            ))}
        </div>
    );
};

export default LanguageList;
