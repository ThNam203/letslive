"use client";

import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import { User } from "../../../../types/user";
import {
    StreamingFrame,
    VideoInfo,
} from "../../../../components/custom_react_player/streaming_frame";
import { GetUserById } from "../../../../lib/api/user";
import ProfileView from "./profile";
import ChatUI from "./chat";
import GLOBAL from "../../../../global";
import { Livestream } from "../../../../types/livestream";
import {
    GetAllLivestreamOfUser,
    IsUserStreaming,
} from "../../../../lib/api/livestream";
import { Button } from "../../../../components/ui/button";
import { Menu } from "lucide-react";

export default function Livestreaming() {
    const [user, setUser] = useState<User | null>(null);
    const [isStreaming, setIsStreaming] = useState(false);
    const [vods, setVods] = useState<Livestream[]>([]); // TODO: change the way this is done, vods cant be livestream
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
        videoTitle: "Live Streaming",
        streamer: {
            name: "Dr. Doppelgangers",
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
                // the first vod is the current livestream
                const { isStreaming, fetchError: streamingError } =
                    await IsUserStreaming(user.id);
                if (streamingError) {
                    toast(streamingError.message, {
                        toastId: "streaming-fetch-error",
                        type: "error",
                    });
                    return;
                }

                setIsStreaming(isStreaming);

                const { livestreams, fetchError: vodsError } =
                    await GetAllLivestreamOfUser(user.id);
                if (vodsError) {
                    toast(vodsError.message, {
                        toastId: "vods-fetch-error",
                        type: "error",
                    });
                    return;
                }

                if (
                    isStreaming &&
                    livestreams.length > 0 &&
                    livestreams[0].status == "live"
                ) {
                    setPlayerInfo({
                        videoTitle: livestreams[0].title,
                        streamer: {
                            name: user.displayName ?? user.username,
                        },
                        videoUrl: `${GLOBAL.API_URL}/transcode/${livestreams[0].id}/index.m3u8`,
                    });
                }

                setVods(livestreams);
            }
        };

        fetchUser();
    }, [params.userId]);

    return (
        <div className="flex h-full overflow-hidden ml-4 gap-6">
            {/* Main content area */}
            <div className="flex-1 overflow-auto no-scrollbar">
                {isStreaming ? (
                    <div className="w-full aspect-video bg-black mb-4 rounded-sm">
                        <StreamingFrame
                            videoInfo={playerInfo}
                            onVideoStart={() => {
                                setTimeVideoStart(new Date());
                            }}
                        />
                    </div>
                ) : (
                    <div className="w-full aspect-video mb-4 bg-black flex items-center justify-center bg-opacity-9 0">
                        <h2 className="text-gray-400 text-3xl font-mono ">
                            The user is currently offline.
                        </h2>
                    </div>
                )}

                {user && <ProfileView user={user} updateUser={updateUser} vods={vods.filter(vod => vod.status !== "live")} />}
            </div>
            {/* Mobile chat toggle button */}
            <Button
                variant="outline"
                size="icon"
                className="fixed bottom-4 right-4 z-50 md:hidden"
                onClick={() => setIsChatOpen(!isChatOpen)}
            >
                <Menu className="h-5 w-5" />
            </Button>

            {/* Chat panel - hidden on mobile unless toggled */}
            <div
                className={`w-full h-full md:w-80 lg:w-96 bg-background transition-all duration-300 fixed md:relative top-0 right-2 z-40 ${isChatOpen ? "translate-x-0" : "translate-x-full md:translate-x-0"}`}
            >
                <ChatUI roomId={params.userId} onClose={() => setIsChatOpen(false)} />
            </div>
        </div>
    );
}
