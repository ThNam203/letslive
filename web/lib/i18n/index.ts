import { createI18nInstance } from "@/lib/i18n/i18next";

export async function myGetT(
    lng: string,
    ns: string | string[] = "common",
) {
    const instance = await createI18nInstance(lng, ns);

    return {
        t: instance.getFixedT(lng, Array.isArray(ns) ? ns[0] : ns),
        i18n: instance,
    };
}
