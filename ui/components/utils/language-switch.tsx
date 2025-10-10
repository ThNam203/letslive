"use client";

import {
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { cn } from "@/utils/cn";
import useT from "@/hooks/use-translation";
import { I18N_COOKIE_NAME, I18N_FALLBACK_LNG, I18N_LANGUAGE_COUNTRY_MAP, I18N_LANGUAGES } from "@/lib/i18n/settings";
import { usePathname, useRouter } from "next/navigation";

const LanguageSwitch = ({ className }: { className?: string }) => {
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
        <Select onValueChange={handleChange} value={i18n.resolvedLanguage}>
            <SelectTrigger className={cn("w-fit border-border", className)}>
                <SelectValue defaultValue={i18n.resolvedLanguage} />
            </SelectTrigger>
            <SelectContent className="border border-border bg-background text-foreground">
                <SelectGroup>
                    {Object.values(I18N_LANGUAGES).map((lng) => (
                        <SelectItem key={lng} value={lng}>
                            {I18N_LANGUAGE_COUNTRY_MAP[lng]}
                        </SelectItem>
                    ))}
                </SelectGroup>
            </SelectContent>
        </Select>
    );
};

export default LanguageSwitch;
