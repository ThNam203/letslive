import i18next, { type i18n } from "i18next";
import resourcesToBackend from "i18next-resources-to-backend";
import { initReactI18next } from "react-i18next/initReactI18next";
import {
    I18N_FALLBACK_LNG,
    I18N_LANGUAGES,
    I18N_DEFAULT_NS,
    I18N_PRELOADED_NS,
    isSupportedLocale,
} from "./settings";

const runsOnServerSide = typeof window === "undefined";

// The root layout renders <html lang> from the resolved locale, so the
// browser starts in the same language the server rendered and hydration
// matches.
function initialLanguage(): string {
    if (runsOnServerSide) return I18N_FALLBACK_LNG;
    const htmlLang = document.documentElement.lang;
    return isSupportedLocale(htmlLang) ? htmlLang : I18N_FALLBACK_LNG;
}

i18next
    .use(initReactI18next)
    .use(
        resourcesToBackend(
            (language: string, namespace: string) =>
                import(`./locales/${language}/${namespace}.json`),
        ),
    )
    .init({
        supportedLngs: I18N_LANGUAGES,
        fallbackLng: I18N_FALLBACK_LNG,
        lng: initialLanguage(),
        ns: I18N_PRELOADED_NS,
        fallbackNS: I18N_DEFAULT_NS,
        defaultNS: I18N_DEFAULT_NS,
    });

const serverInstances = new Map<string, i18n>();

/**
 * On the server one process renders requests in different languages at once,
 * so the shared instance must never switch language there. Each locale gets
 * its own clone instead; clones share the loaded resources.
 */
export function getI18nForLocale(lng: string): i18n {
    if (!runsOnServerSide) return i18next;

    const cached = serverInstances.get(lng);
    if (cached) return cached;
    const clone = i18next.cloneInstance({ lng });
    serverInstances.set(lng, clone);
    return clone;
}

export default i18next;
