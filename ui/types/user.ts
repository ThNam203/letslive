export type User = {
  id: string;
  username: string;
  bio?: string;
  email: string;
  status: UserStatus;
  phoneNumber?: string;
  authProvider: AuthProvider;
  createdAt: string;
  streamAPIKey: string;
  displayName?: string;
  backgroundPicture?: string;
  profilePicture?: string;
  followerCount: number;
  livestreamInformation: LivestreamInformation;
  socialMediaLinks?: Record<string, string>;

  isFollowing?: boolean; // for checking if the current user is following this user
};

export type LivestreamInformation = {
  userId: string;
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
