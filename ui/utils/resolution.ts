/**
 * Extracts the height (vertical resolution) from a resolution string.
 * Handles formats like "1920x1080", "1280x720", etc.
 * Returns the height in pixels, or null if the format is invalid.
 */
export const getResolutionHeight = (resolution: string): number | null => {
    if (!resolution || resolution === "Auto") return null;
    
    // Handle format like "1920x1080"
    const match = resolution.match(/^(\d+)x(\d+)$/);
    if (match) {
        return parseInt(match[2], 10);
    }
    
    // Fallback: try to extract number if it's already in "1080p" format
    const pMatch = resolution.match(/^(\d+)p?$/i);
    if (pMatch) {
        return parseInt(pMatch[1], 10);
    }
    
    return null;
};

/**
 * Formats a resolution string for display (e.g., "1920x1080" -> "1080p")
 * Returns null if the resolution cannot be parsed.
 */
export const formatResolutionForDisplay = (resolution: string): string | null => {
    if (resolution === "Auto") return "Auto";
    
    const height = getResolutionHeight(resolution);
    if (height !== null) {
        return `${height}p`;
    } 

    return null;
};

