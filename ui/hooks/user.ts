import { GetMeProfile, UpdateProfile } from "@/lib/api/user";
import { User } from "@/types/user";
import { create } from "zustand";

export type UserState = {
    user: User | null;
    fetchUser: () => Promise<void>;
    clearUser: () => void;
    updateUser: (user: User) => void;
};

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
    updateUser: (updateUser: User) =>
        set((prev) => {
            console.log("THE DATA AFTER UPDATE", updateUser);
            return {
                user: {
                    ...prev.user,
                    ...updateUser,
                },
            };
        }),
}));

export default useUser;
