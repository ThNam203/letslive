export const SEARCH_PATH = "/search";

export const SEARCH_TYPES = ["users", "lives", "vods"] as const;
export type SearchType = (typeof SEARCH_TYPES)[number];

export const SEARCH_TABS = ["all", ...SEARCH_TYPES] as const;
export type SearchTab = (typeof SEARCH_TABS)[number];

export const SEARCH_TYPING_DEBOUNCE_MS = 300;

// i18next HTML-escapes interpolated values and React escapes them again, so a
// query like "rock & roll" would show as "rock &amp; roll".
export const RAW_INTERPOLATION = {
    interpolation: { escapeValue: false },
} as const;

export function parseSearchTab(value: string | null): SearchTab {
    return SEARCH_TABS.find((tab) => tab === value) ?? "all";
}

export function searchHref(query: string, tab: SearchTab = "all") {
    const params = new URLSearchParams();
    if (query) params.set("q", query);
    if (tab !== "all") params.set("type", tab);
    const qs = params.toString();
    return qs ? `${SEARCH_PATH}?${qs}` : SEARCH_PATH;
}
