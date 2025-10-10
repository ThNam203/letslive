"use client";

import { cn } from "@/utils/cn";
import useT from "@/hooks/use-translation";
import {
    I18N_COOKIE_NAME,
    I18N_FALLBACK_LNG,
    I18N_LANGUAGE_COUNTRY_MAP,
    I18N_LANGUAGES,
} from "@/lib/i18n/settings";
import { usePathname, useRouter } from "next/navigation";

const LanguageList = ({ className }: { className?: string }) => {
    const { i18n } = useT();
    const router = useRouter();
    const pathname = usePathname();

    const handleChange = async (option: string) => {
        const oldLanguage = String(i18n.resolvedLanguage);
        await i18n.changeLanguage(option || I18N_FALLBACK_LNG).then(() => {
            document.cookie = `${I18N_COOKIE_NAME}=${option}; path=/; max-age=${30 * 24 * 60 * 60}`;
            const newPath = pathname.replace(`/${oldLanguage}/`, `/${option}/`);
            router.replace(newPath);
        });
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
                            ? "font-bold ring-2 ring-border"
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
