import { create } from "zustand";
import { User } from "../types/user";

export type UserState = {
    user: User | null;
    setUser: (user: User | null) => void;
    clearUser: () => void;
    updateUser: (user: User) => void;
};

const useUser = create<UserState>((set) => ({
    user: null,
    setUser: (user: User | null) => {
        set({ user });
    },
    clearUser: () => {
        set({ user: null });
    },
    updateUser: (updateUser: User) =>
        set((prev) => {
            return {
                user: {
                    ...prev.user,
                    ...updateUser,
                },
            };
        }),
}));

export default useUser;
