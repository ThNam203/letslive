"use client";

import { useParams } from "next/navigation";
import { useMemo, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { PublicUser } from "@/types/user";
import {
    StreamingFrame,
    VideoInfo,
} from "@/components/custom_react_player/streaming-frame";
import ProfileView from "./profile";
import ChatUI from "./chat";
import GLOBAL from "@/global";
import { Button } from "@/components/ui/button";
import IconMenu from "@/components/icons/menu";
import useT from "@/hooks/use-translation";
import { publicUserQueryKey, usePublicUser } from "@/hooks/queries/use-users";
import { useLivestreamOfUser } from "@/hooks/queries/use-livestream-of-user";
import { usePublicVodsOfUser } from "@/hooks/queries/use-vods";

export default function Livestreaming() {
    const { t } = useT(["common", "users", "fetch-error"]);
    const params = useParams<{ userId: string }>();
    const queryClient = useQueryClient();
    const [isChatOpen, setIsChatOpen] = useState(false);
    const [timeVideoStart, setTimeVideoStart] = useState<Date>(new Date());

    const { data: user } = usePublicUser(params.userId);
    const { data: livestream } = useLivestreamOfUser(params.userId);
    const { data: vods } = usePublicVodsOfUser(params.userId);

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
                      videoTitle: livestream.title,
                      streamer: { name: user?.username ?? "" },
                      videoUrl: `${GLOBAL.API_URL}/transcode/${livestream.id}/index.m3u8`,
                  }
                : {
                      videoTitle: t("common:live_streaming"),
                      streamer: { name: "" },
                      videoUrl: null,
                  },
        [livestream, user?.username, t],
    );

    return (
        <div className="ml-4 flex h-full gap-6 overflow-hidden">
            {/* Main content area */}
            <div className="no-scrollbar flex-1 overflow-auto">
                {livestream ? (
                    <StreamingFrame
                        videoInfo={playerInfo}
                        onVideoStart={() => {
                            setTimeVideoStart(new Date());
                        }}
                        className="mt-1"
                    />
                ) : (
                    <div className="bg-opacity-9 0 mt-1 mb-4 flex aspect-video w-full items-center justify-center bg-black">
                        <h2 className="text-foreground-muted font-mono text-3xl">
                            {t("users:offline")}
                        </h2>
                    </div>
                )}
                {user && (
                    <ProfileView
                        user={user}
                        updateUser={updateUser}
                        vods={vods ?? []}
                        className="mt-2"
                    />
                )}
            </div>
            {/* Mobile chat toggle button */}
            <Button
                variant="outline"
                size="icon"
                className="fixed right-4 bottom-4 z-50 md:hidden"
                onClick={() => setIsChatOpen(!isChatOpen)}
            >
                <IconMenu className="h-5 w-5" />
            </Button>

            {/* Chat panel - hidden on mobile unless toggled */}
            <div
                className={`bg-background fixed top-0 right-2 z-40 h-full w-full transition-all duration-300 md:relative md:w-80 lg:w-96 ${
                    isChatOpen
                        ? "translate-x-0"
                        : "translate-x-full md:translate-x-0"
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
