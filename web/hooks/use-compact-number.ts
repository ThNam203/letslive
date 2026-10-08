import { useMemo } from "react";
import useT from "@/hooks/use-translation";

/** Formats counts in the active language, e.g. 1.2K / 1,2 N. */
export default function useCompactNumber(): (value: number) => string {
    const { i18n } = useT("common");
    const language = i18n.resolvedLanguage;
    const formatter = useMemo(
        () => new Intl.NumberFormat(language, { notation: "compact" }),
        [language],
    );
    return (value) => formatter.format(value);
}
