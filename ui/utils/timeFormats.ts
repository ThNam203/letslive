import { TFunction } from "next-i18next";

export function dateDiffFromNow(pastDate: string, t: TFunction) {
    const now = new Date();
    const past = new Date(pastDate);
    const seconds = Math.round((now.getTime() - past.getTime()) / 1000);

    if (seconds < 60) {
        return t("time.seconds_ago", { count: seconds });
    }

    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) {
        return t("time.minutes_ago", { count: minutes });
    }

    const hours = Math.floor(minutes / 60);
    if (hours < 24) {
        return t("time.hours_ago", { count: hours });
    }

    const days = Math.floor(hours / 24);
    if (days < 7) {
        return t("time.days_ago", { count: days });
    }

    const weeks = Math.floor(days / 7);
    if (days < 30) {
        return t("time.weeks_ago", { count: weeks });
    }

    const months = Math.floor(days / 30);
    if (days < 365) {
        return t("time.months_ago", { count: months });
    }

    const years = Math.floor(days / 365);
    return t("time.years_ago", { count: years });
}

export function formatSeconds(duration: number): string {
    const hours = Math.floor(duration / 3600);
    const minutes = Math.floor((duration % 3600) / 60);
    const seconds = duration % 60;

    return `${hours ? `${hours}:` : ""}${minutes}:${seconds < 10 ? "0" + seconds : seconds}`;
}
