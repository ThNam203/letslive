import { GetMeProfile, UpdateProfile } from "@/lib/api/user";
import { User } from "@/types/user";
import { create } from "zustand";

export type UserState = {
    user: User | null,
    fetchUser: () => Promise<void>,
    clearUser: () => void,
    fetchUpdateUser: (user: User) => Promise<void>,
    updateUser: (user: User) => void,
}

const useUser = create<UserState>((set) => ({
    user: null,
    fetchUser: async () => {
        const response = await GetMeProfile();
        if (response.fetchError) {
            throw response.fetchError;
        }

        set({ user: response.user });
    },
    clearUser: () => {
        set({ user: null });
    },
    fetchUpdateUser: async (updatedUser: User) => {
        const response = await UpdateProfile(updatedUser);
        if (response.fetchError) {
            throw response.fetchError;
        }
        
        set({ user: updatedUser });
    },
    updateUser: (user: User) => set((prev) => ({ ...prev, ...user })),
}));

export default useUser;