import useUser from "@/hooks/user";
import { UpdateProfile } from "@/lib/api/user";
import { I18N_COOKIE_NAME, I18N_FALLBACK_LNG } from "./settings";
import i18next from "./i18next";

type LocaleRouter = {
    replace: (href: string) => void;
};

type SwitchLocaleOptions = {
    /** Write the new locale to the user's DB record. Skip when hydrating FE from a DB value that's already there. */
    syncToBackend?: boolean;
};

export async function switchLocale(
    router: LocaleRouter,
    pathname: string,
    newLocale: string,
    { syncToBackend = true }: SwitchLocaleOptions = {},
): Promise<void> {
    const locale = newLocale || I18N_FALLBACK_LNG;

    await i18next.changeLanguage(locale);

    document.cookie = `${I18N_COOKIE_NAME}=${locale}; path=/; max-age=${30 * 24 * 60 * 60}`;

    const segments = pathname.split("/");
    segments[1] = locale;
    router.replace(segments.join("/"));

    if (syncToBackend && useUser.getState().user) {
        UpdateProfile({ locale }).catch((err) => {
            console.warn("failed to sync locale preference", err);
        });
    }
}
