/**
 * Cuts text to at most `max` UTF-16 units, the way `maxLength` counts, without
 * splitting an emoji: a high surrogate left at the cut is dropped too.
 */
export function truncateText(text: string, max: number): string {
    if (text.length <= max) return text;
    const cut = text.slice(0, max);
    return /[\uD800-\uDBFF]$/.test(cut) ? cut.slice(0, -1) : cut;
}
