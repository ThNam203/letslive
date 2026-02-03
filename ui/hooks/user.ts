import { create } from "zustand";
import { MeUser } from "../types/user";

export type UserState = {
    user: MeUser | null;
    isLoading: boolean;
    setUser: (user: MeUser | null) => void;
    clearUser: () => void;
    updateUser: (user: Partial<MeUser>) => void;
    setIsLoading: (isLoading: boolean) => void;
};

const useUser = create<UserState>((set) => ({
    user: null,
    isLoading: false,

    setUser: (user) => set({ user }),
    clearUser: () => set({ user: null }),
    updateUser: (update) =>
        set((prev) => ({
            user: prev.user ? { ...prev.user, ...update } : null,
        })),
    setIsLoading: (isLoading) => set({ isLoading }),
}));

export default useUser;
