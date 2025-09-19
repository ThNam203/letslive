"use client";

import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import { User } from "@/types/user";
import {
    StreamingFrame,
    VideoInfo,
} from "@/components/custom_react_player/streaming-frame";
import { GetUserById } from "@/lib/api/user";
import ProfileView from "./profile";
import ChatUI from "./chat";
import GLOBAL from "@/global";
import { GetLivestreamOfUser } from "@/lib/api/livestream";
import { Button } from "@/components/ui/button";
import IconMenu from "@/components/icons/menu";
import { VOD } from "@/types/vod";
import { Livestream } from "@/types/livestream";
import { GetPublicVODsOfUser } from "@/lib/api/vod";
import useT from "@/hooks/use-translation";

export default function Livestreaming() {
    const { t } = useT("translation");
    const [user, setUser] = useState<User | null>(null);
    const [livestream, setLivestream] = useState<Livestream | null>(null);
    const [vods, setVods] = useState<VOD[]>([]);
    const [isChatOpen, setIsChatOpen] = useState(false);

    const updateUser = (newUserInfo: User) => {
        setUser((prev) => {
            if (prev) {
                return {
                    ...prev,
                    ...newUserInfo,
                };
            }
            return prev;
        });
    };

    const params = useParams<{ userId: string }>();
    const [playerInfo, setPlayerInfo] = useState<VideoInfo>({
        videoTitle: t("users.live_streaming"),
        streamer: {
            name: "",
        },
        videoUrl: null,
    });

    const [timeVideoStart, setTimeVideoStart] = useState<Date>(new Date());

    useEffect(() => {
        const fetchUser = async () => {
            const { user, fetchError } = await GetUserById(params.userId);
            if (fetchError) {
                toast(fetchError.message, {
                    toastId: "user-fetch-error",
                    type: "error",
                });
                return;
            }

            if (user) {
                setUser(user);
                const { livestream, fetchError: streamingError } =
                    await GetLivestreamOfUser(user.id);
                if (streamingError && streamingError.status !== 404) {
                    // If the error is not a 404, it means there was an issue fetching the livestream
                    toast(streamingError.message, {
                        toastId: "streaming-fetch-error",
                        type: "error",
                    });
                    return;
                }

                if (livestream) {
                    setLivestream(livestream);

                    setPlayerInfo({
                        videoTitle: livestream.title,
                        streamer: {
                            name: user.displayName ?? user.username,
                        },
                        videoUrl: `${GLOBAL.API_URL}/transcode/${livestream.id}/index.m3u8`,
                    });
                }

                const { vods, fetchError: vodsError } =
                    await GetPublicVODsOfUser(user.id);
                if (vodsError) {
                    toast(vodsError.message, {
                        toastId: "vods-fetch-error",
                        type: "error",
                    });
                    return;
                } else setVods(vods);
            }
        };

        fetchUser();
    }, [params.userId]);

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
                    <div className="bg-opacity-9 0 mb-4 flex aspect-video w-full items-center justify-center bg-black mt-1">
                        <h2 className="font-mono text-3xl text-foreground-muted">
                            {t("users.offline")}
                        </h2>
                    </div>
                )}
                {user && (
                    <ProfileView
                        user={user}
                        updateUser={updateUser}
                        vods={vods}
                        className="mt-2"
                    />
                )}
            </div>
            {/* Mobile chat toggle button */}
            <Button
                variant="outline"
                size="icon"
                className="fixed bottom-4 right-4 z-50 md:hidden"
                onClick={() => setIsChatOpen(!isChatOpen)}
            >
                <IconMenu className="h-5 w-5" />
            </Button>

            {/* Chat panel - hidden on mobile unless toggled */}
            <div
                className={`fixed right-2 top-0 z-40 h-full w-full bg-background transition-all duration-300 md:relative md:w-80 lg:w-96 ${
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
