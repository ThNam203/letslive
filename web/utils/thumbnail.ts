import GLOBAL from "@/global";

// Media without an uploaded thumbnail falls back to the frame the transcoder
// captured, stored under the livestream/VOD id.
export function mediaThumbnailUrl(id: string, thumbnailUrl: string | null) {
    return (
        thumbnailUrl ??
        `${GLOBAL.API_URL}/files/livestreams/${id}/thumbnail.jpeg`
    );
}
