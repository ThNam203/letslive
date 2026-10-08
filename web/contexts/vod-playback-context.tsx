"use client";

import { createContext, useContext } from "react";

type VODPlayback = {
    seekTo: (seconds: number) => void;
    /** video length in seconds, 0 when unknown */
    duration: number;
};

const VODPlaybackContext = createContext<VODPlayback | null>(null);

export const VODPlaybackProvider = VODPlaybackContext.Provider;

// null outside a VOD page, where timestamps render as plain text
export function useVODPlayback() {
    return useContext(VODPlaybackContext);
}
