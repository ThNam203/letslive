"use client";

import { useParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { PublicUser } from "@/types/user";
import {
    StreamingFrame,
    VideoInfo,
} from "@/components/custom_react_player/streaming-frame";
import ProfileView from "./profile";
import ProfileSkeleton from "./profile-skeleton";
import { PlayerSkeleton } from "@/components/skeletons/player-skeleton";
import ChatUI from "./chat";
import GLOBAL from "@/global";
import { Button } from "@/components/ui/button";
import IconMessage from "@/components/icons/message";
import useT from "@/hooks/use-translation";
import { publicUserQueryKey, usePublicUser } from "@/hooks/queries/use-users";
import { useLivestreamOfUser } from "@/hooks/queries/use-livestream-of-user";
import IconRefresh from "@/components/icons/refresh";

export default function Livestreaming() {
    const { t } = useT(["common", "users", "fetch-error", "chat"]);
    const params = useParams<{ userId: string }>();
    const queryClient = useQueryClient();
    const [isChatOpen, setIsChatOpen] = useState(false);
    const [timeVideoStart, setTimeVideoStart] = useState<Date>(new Date());

    const {
        data: user,
        isLoading: isLoadingUser,
        isError: isUserError,
        refetch: refetchUser,
    } = usePublicUser(params.userId);
    const { data: livestream, isLoading: isLoadingLivestream } =
        useLivestreamOfUser(params.userId);

    useEffect(() => {
        if (!isChatOpen) return;
        const closeOnEscape = (e: KeyboardEvent) => {
            if (e.key === "Escape") setIsChatOpen(false);
        };
        window.addEventListener("keydown", closeOnEscape);
        return () => window.removeEventListener("keydown", closeOnEscape);
    }, [isChatOpen]);

    // ProfileView edits the profile in place (follow, gift, socials), so the
    // cached copy is patched rather than refetched.
    const updateUser = (newUserInfo: PublicUser) => {
        queryClient.setQueryData<PublicUser>(
            publicUserQueryKey(params.userId),
            (prev) => (prev ? { ...prev, ...newUserInfo } : prev),
        );
    };

    const playerInfo: VideoInfo = useMemo(
        () =>
            livestream
                ? {
                      videoUrl: `${GLOBAL.API_URL}/transcode/${livestream.id}/index.m3u8`,
                  }
                : {
                      videoUrl: null,
                  },
        [livestream],
    );

    return (
        <div className="flex h-full overflow-hidden xl:gap-6">
            {/* Main content area */}
            <div className="no-scrollbar min-w-0 flex-1 overflow-auto px-3 pb-24 sm:px-4 xl:pr-0 xl:pb-10">
                {/* Player only while live; offline channels show a banner instead. */}
                {isLoadingLivestream ? (
                    <PlayerSkeleton className="mt-1" />
                ) : (
                    livestream && (
                        <StreamingFrame
                            videoInfo={playerInfo}
                            onVideoStart={() => {
                                setTimeVideoStart(new Date());
                            }}
                            className="mt-1"
                        />
                    )
                )}
                {isLoadingUser || isLoadingLivestream ? (
                    <ProfileSkeleton />
                ) : isUserError ? (
                    <div
                        role="alert"
                        className="mt-10 flex flex-col items-center gap-3 text-center"
                    >
                        <p>{t("fetch-error:client_fetch_error")}</p>
                        <Button variant="outline" onClick={() => refetchUser()}>
                            <IconRefresh />
                            {t("users:channel.retry")}
                        </Button>
                    </div>
                ) : (
                    user && (
                        <ProfileView
                            user={user}
                            livestream={livestream ?? null}
                            updateUser={updateUser}
                        />
                    )
                )}
            </div>
            {/* Below xl, chat is a slide-over so it doesn't squeeze the channel. */}
            {!isChatOpen && (
                <Button
                    className="fixed right-4 bottom-4 z-50 rounded-full shadow-lg xl:hidden"
                    onClick={() => setIsChatOpen(true)}
                    aria-expanded={false}
                    aria-controls="channel-chat"
                >
                    <IconMessage />
                    {t("chat:title")}
                </Button>
            )}
            {isChatOpen && (
                <div
                    aria-hidden="true"
                    className="fixed inset-0 z-40 bg-black/40 xl:hidden"
                    onClick={() => setIsChatOpen(false)}
                />
            )}

            <div
                id="channel-chat"
                className={`bg-background fixed inset-y-0 right-0 z-50 w-full shadow-xl transition-transform duration-300 sm:w-96 xl:relative xl:inset-auto xl:z-auto xl:w-96 xl:shrink-0 xl:shadow-none ${
                    isChatOpen
                        ? "translate-x-0"
                        : "translate-x-full xl:translate-x-0"
                }`}
            >
                <ChatUI
                    roomId={params.userId}
                    onClose={() => setIsChatOpen(false)}
                />
            </div>
        </div>
    );
}
