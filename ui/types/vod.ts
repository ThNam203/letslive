export type VOD = {
    id: string;
    livestreamId: string;
    userId: string;
    title: string;
    description: string | null;
    thumbnailUrl: string | null;
    visibility: "public" | "private";
    viewCount: number;
    duration: number;
    playbackUrl: string;
    createdAt: string; // ISO 8601 timestamp
    updatedAt: string; // ISO 8601 timestamp
};
