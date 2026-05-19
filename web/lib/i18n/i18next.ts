import i18next, { createInstance, i18n as I18nInstance } from "i18next";
import resourcesToBackend from "i18next-resources-to-backend";
import { initReactI18next } from "react-i18next/initReactI18next";
import {
    I18N_FALLBACK_LNG,
    I18N_LANGUAGES,
    I18N_DEFAULT_NS,
    I18N_COOKIE_NAME,
} from "./settings";
import LanguageDetector from "i18next-browser-languagedetector";

const runsOnServerSide = typeof window === "undefined";

const loadResources = resourcesToBackend(
    (language: string, namespace: string) =>
        import(`./locales/${language}/${namespace}.json`),
);

if (!runsOnServerSide) {
    i18next
        .use(initReactI18next)
        .use(LanguageDetector)
        .use(loadResources)
        .init({
            supportedLngs: I18N_LANGUAGES,
            fallbackLng: I18N_FALLBACK_LNG,
            lng: undefined,
            fallbackNS: I18N_DEFAULT_NS,
            defaultNS: I18N_DEFAULT_NS,
            detection: {
                order: ["path", "cookie", "navigator", "htmlTag"],
                lookupCookie: I18N_COOKIE_NAME,
            },
        });
}

export async function createI18nInstance(
    lng: string,
    ns: string | string[] = I18N_DEFAULT_NS,
): Promise<I18nInstance> {
    const instance = createInstance();
    await instance.use(initReactI18next).use(loadResources).init({
        supportedLngs: I18N_LANGUAGES,
        fallbackLng: I18N_FALLBACK_LNG,
        lng,
        fallbackNS: I18N_DEFAULT_NS,
        defaultNS: I18N_DEFAULT_NS,
        ns,
    });
    return instance;
}

export default i18next;
