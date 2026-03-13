export type Livestream = {
    id: string;
    userId: string;
    title: string;
    description: string | null;
    thumbnailUrl: string | null;
    viewCount: number;
    visibility: "public" | "private";
    startedAt: string;
    endedAt: string | null;
    createdAt: string;
    updatedAt: string;
    vodId: string | null;
};
