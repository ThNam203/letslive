export type User = {
    id: string;
    username: string;
    bio?: string;
    email: string;
    isOnline: boolean;
    isVerified: boolean;
    isActive: boolean;
    createdAt: string;
    streamAPIKey: string;
    vods: UserVOD[] | null;
    displayName?: string;
    backgroundPicture?: string;
    profilePicture?: string;
    livestreamInformation: LivestreamInformation;
};

export type LivestreamInformation = {
    userId: string;
    title: string | null;
    description: string | null;
    thumbnailUrl: string | null;
};

export type UserVOD = {
    id: string;
    userId: string;
    title: string;
    description: string;
    thumbnailUrl: string;
    status: string;
    viewCount: number;
    startedAt: string;
    endedAt: string;
    playbackUrl: string;
    createdAt: string;
    updatedAt: string;
};
