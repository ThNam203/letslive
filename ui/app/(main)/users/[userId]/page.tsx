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

export default function Livestreaming() {
    const [user, setUser] = useState<User | null>(null);
    const [isStreaming, setIsStreaming] = useState(false);
    const [vods, setVods] = useState<Livestream[]>([]); // TODO: change the way this is done, vods cant be livestream

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
        <div className="overflow-y-auto h-full flex lg:flex-row max-lg:flex-col mt-2">
            <div className="w-[1200px] min-w-[1200px]">
                {isStreaming ? (
                    <>
                        <div className="w-full h-[675px] bg-black mb-4 rounded-sm">
                            <StreamingFrame
                                videoInfo={playerInfo}
                                onVideoStart={() => {
                                    setTimeVideoStart(new Date());
                                }}
                            />
                        </div>
                        {/* <div className="w-full font-sans my-4 overflow-x-auto whitespace-nowrap">
                            {servers.map((_, idx) => (
                                <Button
                                    key={idx}
                                    onClick={() => setServerIndex(idx)}
                                    className={cn(
                                        "mr-4",
                                        serverIndex == idx ? "bg-green-700" : ""
                                    )}
                                >
                                    Server {idx + 1}
                                </Button>
                            ))}
                        </div> */}
                    </>
                ) : (
                    <div className="w-full h-[675px] mb-4 bg-black flex items-center justify-center bg-opacity-9 0">
                        <h2 className="text-gray-400 text-3xl font-mono ">
                            The user is currently offline.
                        </h2>
                    </div>
                )}

                {user && <ProfileView user={user} updateUser={updateUser} vods={vods.filter(vod => vod.status !== "live")} />}
            </div>
            <div className="w-[400px] mx-4 fixed right-0 top-12 bottom-4">
                <ChatUI roomId={params.userId} />
            </div>
        </div>
    );
}
