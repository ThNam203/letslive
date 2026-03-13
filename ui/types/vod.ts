export type VODStatus = "uploading" | "processing" | "ready" | "failed";

export type VOD = {
    id: string;
    livestreamId: string | null;
    userId: string;
    title: string;
    description: string | null;
    thumbnailUrl: string | null;
    visibility: "public" | "private";
    viewCount: number;
    duration: number;
    playbackUrl: string | null;
    status: VODStatus;
    originalFileUrl: string | null;
    createdAt: string; // ISO 8601 timestamp
    updatedAt: string; // ISO 8601 timestamp
};
