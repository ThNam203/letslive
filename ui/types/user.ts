export type User = {
  id: string;
  username: string;
  bio: string;
  email: string;
  isOnline: boolean;
  isVerified: boolean;
  createdAt: string;
  streamAPIKey: string;
  vods: string[];
};