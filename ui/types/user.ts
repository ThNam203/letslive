export type User = {
    id: string;
    username: string;
    bio?: string;
    email: string;
    liveStatus: UserLiveStatus;
    status: UserStatus;
    authProvider: AuthProvider;
    isVerified: boolean;
    createdAt: string;
    streamAPIKey: string;
    vods: UserVOD[] | null;
    displayName?: string;
    backgroundPicture?: string;
    profilePicture?: string;
    followerCount: number;
    livestreamInformation: LivestreamInformation;

    isFollowing?: boolean; // for checking if the current user is following this user
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
    description: string | null;
    thumbnailUrl: string;
    status: string;
    viewCount: number;
    startedAt: string;
    endedAt: string;
    playbackUrl: string;
    createdAt: string;
    updatedAt: string;
};

export enum UserLiveStatus {
    LIVE = 'on',
    OFFLIVE = 'off',
}

export enum AuthProvider {
    GOOGLE = 'google',
    LOCAL = 'local',
}

export enum UserStatus {
    NORMAL = 'normal',
    DISABLED = 'disabled',
}
