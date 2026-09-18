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
import { I18N_LANGUAGE_COUNTRY_MAP, I18N_LANGUAGES } from "@/lib/i18n/settings";
import { switchLocale } from "@/lib/i18n/switch-locale";
import { usePathname, useRouter } from "next/navigation";
import { useUpdateProfile } from "@/hooks/queries/use-profile-mutations";

const LanguageSwitch = ({ className }: { className?: string }) => {
    const { i18n } = useT();
    const router = useRouter();
    const pathname = usePathname();
    const updateProfile = useUpdateProfile();

    const handleChange = async (option: string) => {
        await switchLocale(router, pathname, option, {
            syncLocale: (locale) => updateProfile.mutate({ locale }),
        });
    };

    return (
        <Select onValueChange={handleChange} value={i18n.resolvedLanguage}>
            <SelectTrigger className={cn("border-border w-fit", className)}>
                <SelectValue defaultValue={i18n.resolvedLanguage} />
            </SelectTrigger>
            <SelectContent className="border-border bg-background text-foreground border">
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
