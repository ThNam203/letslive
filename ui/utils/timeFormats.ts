export function datediffFromNow(pastDate: string) {        
    const now = new Date();
    const past = new Date(pastDate);
    const seconds = Math.round((now.getTime() - past.getTime()) / 1000);

    if (seconds < 60) {
        return `${seconds} second${seconds !== 1 ? 's' : ''}`;
    }

    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) {
        return `${minutes} minute${minutes !== 1 ? 's' : ''}`;
    }

    const hours = Math.floor(minutes / 60);
    if (hours < 24) {
        return `${hours} hour${hours !== 1 ? 's' : ''}`;
    }

    const days = Math.floor(hours / 24);
    if (days < 7) {
        return `${days} day${days !== 1 ? 's' : ''}`;
    }

    const weeks = Math.floor(days / 7);
    if (days < 30) {
        return `${weeks} week${weeks !== 1 ? 's' : ''}`;
    }

    const months = Math.floor(days / 30);
    if (days < 365) {
        return `${months} month${months !== 1 ? 's' : ''}`;
    }

    const years = Math.floor(days / 365);
    return `${years} year${years !== 1 ? 's' : ''}`;
}

export function formatSeconds(duration: number): string {
    const hours = Math.floor(duration / 3600);
    const minutes = Math.floor((duration % 3600) / 60);
    const seconds = duration % 60;

    return `${hours ? `${hours}:` : ''}${minutes}:${seconds < 10 ? '0' + seconds : seconds}`;
}