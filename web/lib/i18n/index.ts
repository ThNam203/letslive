import { headers } from "next/headers";
import { I18N_HEADER_NAME, I18N_FALLBACK_LNG } from "./settings";
import { createI18nInstance } from "@/lib/i18n/i18next";

export async function myGetT(ns: string | string[] = "common") {
    const headerList = await headers();
    const lng = headerList.get(I18N_HEADER_NAME) ?? I18N_FALLBACK_LNG;

    const instance = await createI18nInstance(lng, ns);

    return {
        t: instance.getFixedT(lng, Array.isArray(ns) ? ns[0] : ns),
        i18n: instance,
    };
}
