import { createSlice, PayloadAction } from "@reduxjs/toolkit";

export type AuthState = {
    id: string;
    username: string;
    email: string;
    isVerified: boolean;
};

const authSlice = createSlice({
    name: "auth",
    initialState: null as AuthState | null,
    reducers: {
        setAuthState: (state, action: PayloadAction<AuthState | null>) => {
            return action.payload;
        },
    },
});

export const { setAuthState } = authSlice.actions;
export default authSlice.reducer;
