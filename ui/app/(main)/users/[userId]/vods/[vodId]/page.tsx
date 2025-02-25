"use client";
import ProfileView from "@/app/(main)/users/[userId]/profile";
import {
    StreamingFrame,
    VideoInfo,
} from "@/components/custom_react_player/streaming_frame";
import { VODFrame } from "@/components/custom_react_player/vod_frame";
import { Button } from "@/components/ui/button";
import VODLink from "@/components/vodlink";
import { GetUserById } from "@/lib/api/user";
import { GetVODInformation } from "@/lib/api/vod";
import { cn } from "@/lib/utils";
import { User } from "@/types/user";
import Image from "next/image";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";

export default function VODPage() {
    const [user, setUser] = useState<User | null>(null);
    const params = useParams<{ userId: string; vodId: string }>();

    const [playerInfo, setPlayerInfo] = useState<VideoInfo>({
        videoTitle: "Live Streaming",
        streamer: {
            name: "Dr. Pedophile",
        },
        videoUrl: null,
    });

    useEffect(() => {
        const fetchVODInfo = async () => {
            const { vodInfo, fetchError } = await GetVODInformation(
                params.vodId
            );

            if (fetchError) {
                toast.error(fetchError.message, {
                    toastId: "vod-fetch-error",
                });
                return;
            }

            const newUrl = `http://localhost:8889/static/vods/${params.vodId}/index.m3u8`;
            setPlayerInfo((prev) => ({
                ...prev,
                videoTitle: vodInfo!.title,
                videoUrl: newUrl,
            }));
        };

        fetchVODInfo();
    }, [params.userId, params.vodId]);

    useEffect(() => {
        const fetchUserInfo = async () => {
            const { user, fetchError } = await GetUserById(params.userId);
            if (fetchError != undefined) {
                toast.error(fetchError.message, {
                    toastId: "user-fetch-error",
                });
            } else {
                setUser(user!);
            }
        };

        fetchUserInfo();
    }, [params.userId]);

    return (
        <div className="overflow-y-auto h-full">
            <div className="flex lg:flex-row max-lg:flex-col">
                <div className="w-[1200px] min-w-[1200px]">
                    <div className="w-full h-[675px] bg-black mb-4">
                        <VODFrame videoInfo={playerInfo} />
                    </div>
                    {/* <div className="w-full font-sans mt-4 overflow-x-auto whitespace-nowrap">
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
                {user && <ProfileView user={user} showSavedStream={false}/>}
                </div>

                <div className="w-[350px] mx-4 fixed right-0 top-16  bottom-4">
                    <div className="w-full h-full font-sans border border-gray-200 rounded-md bg-gray-50 p-4">
                        <h2 className="text-xl mb-4 font-bold">OTHER STREAMS</h2>
                        <div className="flex flex-col gap-4">
                            {user &&
                                user.vods
                                    ?.filter(
                                        (v) =>
                                            v.id !== params.vodId &&
                                            v.status !== "live"
                                    )
                                    .map((vod, idx) => (
                                        <VODLink key={idx} item={vod} />
                                    ))}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
