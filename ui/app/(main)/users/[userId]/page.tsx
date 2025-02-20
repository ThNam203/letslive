"use client";
import ChatUI from "@/app/(main)/users/[userId]/chat";
import ProfileView from "@/app/(main)/users/[userId]/profile";
import {
    StreamingFrame,
    VideoInfo,
} from "@/components/custom_react_player/streaming_frame";
import { Button } from "@/components/ui/button";
import useUser from "@/hooks/user";
import { cn } from "@/lib/utils";
import { FetchError } from "@/types/fetch-error";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";

const servers = ["", "localhost:8890", "localhost:8891"];

export default function Livestreaming() {
    const user = useUser((state) => state.user);
    const fetchUser = useUser((state) => state.fetchUser);
    const params = useParams<{ userId: string }>();
    const [serverIndex, setServerIndex] = useState(0);
    const [playerInfo, setPlayerInfo] = useState<VideoInfo>({
        videoTitle: "Live Streaming",
        streamer: {
            name: "Dr. Doppelgangers",
        },
        videoUrl: null,
    });

    const [timeVideoStart, setTimeVideoStart] = useState<Date>(new Date());

    useEffect(() => {
        fetchUser().then(() => {
            if (user?.isOnline == false) return;
            const newUrl =
                serverIndex == 0
                    ? `http://localhost:8889/static/${params.userId}/index.m3u8`
                    : `http://localhost:8889/static/${params.userId}/${servers[serverIndex]}_index.m3u8`;

            setPlayerInfo((prev) => ({
                ...prev,
                videoUrl: newUrl,
            }));
        }).catch((e: FetchError) => {
            toast(e.message, {
                toastId: "user-fetch-error",
                type: "error"
            });
        })
    }, [params.userId]);

    useEffect(() => {
        if (user?.isOnline == false) return;

        const newUrl =
            serverIndex == 0
                ? `http://localhost:8889/static/${params.userId}/index.m3u8`
                : `http://localhost:8889/static/${params.userId}/${servers[serverIndex]}_index.m3u8`;

        setPlayerInfo((prev) => ({
            ...prev,
            videoUrl: newUrl,
        }));
    }, [serverIndex]);

    return (
        <div className="overflow-y-auto h-full flex lg:flex-row max-lg:flex-col">
            <div className="w-[1200px] min-w-[1200px]">
                {user && user.isOnline ? (
                    <>
                        <div className="w-full h-[675px] bg-black">
                            <StreamingFrame
                                videoInfo={playerInfo}
                                onVideoStart={() => {
                                    setTimeVideoStart(new Date());
                                }}
                            />
                        </div>
                        <div className="w-full font-sans my-4 overflow-x-auto whitespace-nowrap">
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
                        </div>
                    </>
                ) : (
                        <div className="w-full h-[675px] mb-4 bg-black flex items-center justify-center bg-opacity-9 0">
                            <h2 className="text-gray-400 text-3xl font-mono ">The user is currently offline.</h2>
                        </div>
                )}

                {user && <ProfileView user={user}/>}
            </div>
            <div className="w-[400px] mx-4 fixed right-0 top-12 bottom-4">
                <ChatUI roomId={params.userId}/>
            </div>
        </div>
    );
}
