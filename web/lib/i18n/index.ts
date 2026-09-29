import { headers } from "next/headers";
import {
    I18N_FALLBACK_LNG,
    I18N_HEADER_NAME,
    isSupportedLocale,
} from "./settings";
import { getI18nForLocale } from "@/lib/i18n/i18next";

export async function getLocale(): Promise<string> {
    const headerList = await headers();
    const lng = headerList.get(I18N_HEADER_NAME);
    return isSupportedLocale(lng) ? lng : I18N_FALLBACK_LNG;
}

export async function myGetT(ns: string | string[] = "common") {
    const lng = await getLocale();
    const i18n = getI18nForLocale(lng);

    // a fresh clone reports the parent's language until its first load ends;
    // each clone only ever targets its own locale, so this cannot race
    if (i18n.language !== lng) {
        await i18n.changeLanguage(lng);
    }
    if (!i18n.hasLoadedNamespace(ns)) {
        await i18n.loadNamespaces(ns);
    }

    return {
        t: i18n.getFixedT(lng, Array.isArray(ns) ? ns[0] : ns),
        i18n,
        lng,
    };
}
