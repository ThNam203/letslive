export const LEFT_BAR_PANEL_ID = "left-bar";
export const LEFT_BAR_COLLAPSED_WIDTH_PX = 64;
export const LEFT_BAR_MIN_WIDTH_PX = 240;
export const LEFT_BAR_DEFAULT_WIDTH_PX = 280;
export const LEFT_BAR_MAX_WIDTH = "35%";

export const LEFT_BAR_COOKIE_NAME = "left-bar";
const LEFT_BAR_COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24 * 365;
const LEFT_BAR_COOKIE_PATTERN = /^(\d{1,5}):([01])$/;

export type LeftBarLayout = {
    width: number;
    isCollapsed: boolean;
};

export const DEFAULT_LEFT_BAR_LAYOUT: LeftBarLayout = {
    width: LEFT_BAR_DEFAULT_WIDTH_PX,
    isCollapsed: false,
};

export function parseLeftBarLayout(value: string | undefined): LeftBarLayout {
    const match = value?.match(LEFT_BAR_COOKIE_PATTERN);
    if (!match) return DEFAULT_LEFT_BAR_LAYOUT;
    return {
        width: Math.max(LEFT_BAR_MIN_WIDTH_PX, Number(match[1])),
        isCollapsed: match[2] === "1",
    };
}

export function saveLeftBarLayout({ width, isCollapsed }: LeftBarLayout) {
    const value = `${Math.round(width)}:${isCollapsed ? 1 : 0}`;
    document.cookie = `${LEFT_BAR_COOKIE_NAME}=${value}; path=/; max-age=${LEFT_BAR_COOKIE_MAX_AGE_SECONDS}; samesite=lax`;
}
