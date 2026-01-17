import { create } from "zustand";
import { User } from "@/types/user";

export type UserState = {
    user: User | null;
    isLoading: boolean;
    setUser: (user: User | null) => void;
    clearUser: () => void;
    updateUser: (user: User) => void;
    setIsLoading: (isLoading: boolean) => void;
};

const useUser = create<UserState>((set) => ({
    user: null,
    isLoading: false,

    setUser: (user) => set({ user }),
    clearUser: () => set({ user: null }),
    updateUser: (updateUser) =>
        set((prev) => ({
            user: prev.user ? { ...prev.user, ...updateUser } : updateUser,
        })),
    setIsLoading: (isLoading) => set({ isLoading }),
}));

export default useUser;
