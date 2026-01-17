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
import ProfileView from "@/routes/[lng]/(main)/users/[userId]/profile";
import ChatUI from "@/routes/[lng]/(main)/users/[userId]/chat";
import GLOBAL from "@/global";
import { GetLivestreamOfUser } from "@/lib/api/livestream";
import { Button } from "@/components/ui/button";
import IconMenu from "@/components/icons/menu";
import { VOD } from "@/types/vod";
import { Livestream } from "@/types/livestream";
import { GetPublicVODsOfUser } from "@/lib/api/vod";
import useT from "@/hooks/use-translation";

export default function Livestreaming() {
    const { t } = useT(["common", "users"]);
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
        videoTitle: t("common:live_streaming"),
        streamer: {
            name: "",
        },
        videoUrl: null,
    });

    const [timeVideoStart, setTimeVideoStart] = useState<Date>(new Date());

    useEffect(() => {
        const fetchUser = async () => {
            try {
                const userRes = await GetUserById(params.userId);

                if (!userRes.success) {
                    toast(t(`api-response:${userRes.key}`), {
                        toastId: userRes.requestId,
                        type: "error",
                    });
                    return;
                }

                if (!userRes.data) throw new Error(); // TODO: throw meaning full error

                setUser(userRes.data);

                const livestreamRes = await GetLivestreamOfUser(
                    userRes.data.id,
                );

                // only care if user is live
                if (livestreamRes.success && livestreamRes.data) {
                    setLivestream(livestreamRes.data);

                    setPlayerInfo({
                        videoTitle: livestreamRes.data.title,
                        streamer: {
                            name:
                                userRes.data!.displayName ??
                                userRes.data!.username,
                        },
                        videoUrl: `${GLOBAL.API_URL}/transcode/${livestreamRes.data.id}/index.m3u8`,
                    });
                }

                const vodsRes = await GetPublicVODsOfUser(userRes.data.id);

                if (!vodsRes.success) {
                    toast(t(`api-response:${vodsRes.key}`), {
                        toastId: vodsRes.requestId,
                        type: "error",
                    });
                } else {
                    setVods(vodsRes.data ?? []);
                }
            } catch (err) {
                console.error(err);
                toast(t("api-response:unexpected_error"), { type: "error" });
            }
        };

        fetchUser();
    }, [params.userId, t]);

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
                    <div className="bg-opacity-9 0 mb-4 mt-1 flex aspect-video w-full items-center justify-center bg-black">
                        <h2 className="text-foreground-muted font-mono text-3xl">
                            {t("users:offline")}
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
                className={`bg-background fixed right-2 top-0 z-40 h-full w-full transition-all duration-300 md:relative md:w-80 lg:w-96 ${
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
