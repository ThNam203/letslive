"use client";
import ChatUI from "@/app/(main)/users/[userId]/chat";
import ProfileView from "@/app/(main)/users/[userId]/profile";
import {
    StreamingFrame,
    VideoInfo,
} from "@/components/custom_react_player/streaming_frame";
import useUser from "@/hooks/user";
import { GetUserById } from "@/lib/api/user";
import { User, UserLiveStatus } from "@/types/user";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";

export default function Livestreaming() {
    const [user, setUser] = useState<User | null>(null);
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
    }

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
                    type: "error"
                });
                return;
            }
            
            if (user) {
                setUser(user);
                if (user.liveStatus === UserLiveStatus.LIVE && user.vods && user.vods.length > 0) {
                    setPlayerInfo({
                        videoTitle: user.vods[0].title,
                        streamer: {
                            name: user.displayName ?? user.username,
                        },
                        videoUrl: `http://localhost:8889/static/${user.vods[0].id}/index.m3u8`,
                    });
                }
            }
        };

        fetchUser();
    }, [params.userId]);

    return (
        <div className="overflow-y-auto h-full flex lg:flex-row max-lg:flex-col mt-2">
            <div className="w-[1200px] min-w-[1200px]">
                {user && user.liveStatus === UserLiveStatus.LIVE ? (
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
                            <h2 className="text-gray-400 text-3xl font-mono ">The user is currently offline.</h2>
                        </div>
                )}

                {user && <ProfileView user={user} updateUser={updateUser}/>}
            </div>
            <div className="w-[400px] mx-4 fixed right-0 top-12 bottom-4">
                <ChatUI roomId={params.userId}/>
            </div>
        </div>
    );
}
