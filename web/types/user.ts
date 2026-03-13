export type SocialMediaLinks = {
    facebook?: string;
    twitter?: string;
    instagram?: string;
    linkedin?: string;
    github?: string;
    youtube?: string;
    website?: string;
    tiktok?: string;
};

export type LivestreamInformation = {
    title: string | null;
    description: string | null;
    thumbnailUrl: string | null;
};

export enum AuthProvider {
    GOOGLE = "google",
    LOCAL = "local",
}

export enum UserStatus {
    NORMAL = "normal",
    DISABLED = "disabled",
}

export type BaseUser = {
    id: string;
    username: string;
    status: UserStatus;
    authProvider: AuthProvider;
    createdAt: string;
    displayName?: string;
    bio?: string;
    backgroundPicture?: string;
    profilePicture?: string;
    followerCount: number;
    livestreamInformation: LivestreamInformation;
    socialMediaLinks?: SocialMediaLinks;
};

export type PublicUser = BaseUser & {
    email: string;

    /** AUTH ONLY - indicates if the current user is following this public user*/
    isFollowing?: boolean;
};

export type MeUser = BaseUser & {
    email: string;
    phoneNumber?: string;
    streamAPIKey: string;
};

/** Union for code that can receive either (e.g. profile header when viewing self vs others). */
export type User = PublicUser | MeUser;
