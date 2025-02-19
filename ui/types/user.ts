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
  vods: string[] | null;
  displayName?: string;
  backgroundPicture?: string;
  profilePicture?: string;
};