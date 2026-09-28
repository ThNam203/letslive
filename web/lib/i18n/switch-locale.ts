import useUser from "@/hooks/user";
import {
    I18N_COOKIE_MAX_AGE_SECONDS,
    I18N_COOKIE_NAME,
    I18N_FALLBACK_LNG,
} from "./settings";
import i18next from "./i18next";

type LocaleRouter = {
    refresh: () => void;
};

type SwitchLocaleOptions = {
    /**
     * Persists the new locale to the user's profile, called only when someone
     * is signed in. This module runs outside React and cannot hold the
     * mutation itself, so the caller passes it in. Omit it when hydrating the
     * frontend from a value that already came from the DB.
     */
    syncLocale?: (locale: string) => void;
};

export async function switchLocale(
    router: LocaleRouter,
    newLocale: string,
    { syncLocale }: SwitchLocaleOptions = {},
): Promise<void> {
    const locale = newLocale || I18N_FALLBACK_LNG;

    document.cookie = `${I18N_COOKIE_NAME}=${locale}; path=/; max-age=${I18N_COOKIE_MAX_AGE_SECONDS}; samesite=lax`;

    await i18next.changeLanguage(locale);

    // server components read the locale from the cookie via middleware
    router.refresh();

    if (syncLocale && useUser.getState().user) {
        syncLocale(locale);
    }
}
