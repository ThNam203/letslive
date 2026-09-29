export const I18N_FALLBACK_LNG = "en-US";
export const I18N_LANGUAGES = [I18N_FALLBACK_LNG, "vi-VN"];
export const I18N_DEFAULT_NS = "translation";
// loaded at startup: error toasts translate these outside React, so they
// must not depend on some mounted component having requested them
export const I18N_PRELOADED_NS = ["common", "api-response", "fetch-error"];
export const I18N_COOKIE_NAME = "lng";
// keep in sync with legal:cookie_policy_row_lng_duration
export const I18N_COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24 * 365;
export const I18N_HEADER_NAME = "X-Letslive-Locale";

export const I18N_LANGUAGE_COUNTRY_MAP: Record<string, string> = {
    "en-US": "English",
    "vi-VN": "Tiếng Việt",
};

export function isSupportedLocale(
    value: string | null | undefined,
): value is string {
    return !!value && I18N_LANGUAGES.includes(value);
}

/**
 * Picks the best supported locale from an Accept-Language header, honouring
 * q-values. Matches the exact tag first, then the primary subtag, since
 * browsers often send "vi" rather than "vi-VN".
 */
export function matchAcceptLanguage(header: string | null): string | null {
    if (!header) return null;

    const ranked = header
        .split(",")
        .map((part) => {
            const [tag, ...params] = part.trim().split(";");
            const q = params.find((p) => p.trim().startsWith("q="));
            const weight = q ? Number(q.trim().slice(2)) : 1;
            return { tag: tag.toLowerCase(), weight };
        })
        .filter((entry) => entry.tag && entry.weight > 0)
        .sort((a, b) => b.weight - a.weight);

    for (const { tag } of ranked) {
        const exact = I18N_LANGUAGES.find((l) => l.toLowerCase() === tag);
        if (exact) return exact;
        const primary = tag.split("-")[0];
        const partial = I18N_LANGUAGES.find(
            (l) => l.split("-")[0].toLowerCase() === primary,
        );
        if (partial) return partial;
    }
    return null;
}
