export const NOTIFICATION_POLL_INTERVAL_MS = 30_000;
export const NOTIFICATION_POPUP_LIMIT = 10;

export type TimeAgoTranslator = (
    key: string,
    opts?: { [k: string]: number },
) => string;

export function timeAgo(dateStr: string, t: TimeAgoTranslator): string {
    const now = Date.now();
    const diff = now - new Date(dateStr).getTime();
    const seconds = Math.floor(diff / 1000);
    if (seconds < 60) return t("notification:time_just_now");
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return t("notification:time_m_ago", { m: minutes });
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return t("notification:time_h_ago", { h: hours });
    const days = Math.floor(hours / 24);
    return t("notification:time_d_ago", { d: days });
}
