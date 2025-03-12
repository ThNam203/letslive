export type Livestream = {
    id: string;
    userId: string;
    title: string;
    description: string | null;
    thumbnailUrl: string;
    status: string;
    viewCount: number;
    visibility: "public" | "private";
    startedAt: string;
    endedAt: string;
    playbackUrl: string;
    createdAt: string;
    updatedAt: string;
    duration: number;
};