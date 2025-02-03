"use client";
import {
    StreamingFrame,
    VideoInfo,
} from "@/components/custom_react_player/streaming_frame";
import { GetUserById } from "@/lib/api/user";
import { User } from "@/types/user";
import Image from "next/image";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";

export default function Livestreaming() {
    const params = useParams<{ userId: string }>();
    const [user, setUser] = useState<User | null>(null);
    const [playerInfo, setPlayerInfo] = useState<VideoInfo>({
        videoTitle: "Live Streaming",
        streamer: {
            name: "Dr. Doppelgangers",
        },
        videoUrl: null,
    });

    const [timeVideoStart, setTimeVideoStart] = useState<Date>(new Date());

    useEffect(() => {
        setPlayerInfo((prev) => ({
            ...prev,
            videoUrl: `http://localhost:8889/static/${params.userId}/index.m3u8`,
        }));

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
        <div className="overflow-y-auto h-full flex lg:flex-row max-lg:flex-col">
            <div className="w-[1200px] min-w-[1200px]">
                <div className="w-full h-[675px] bg-black">
                    <StreamingFrame
                        videoInfo={playerInfo}
                        onVideoStart={() => {
                            setTimeVideoStart(new Date());
                        }}
                    />
                </div>
                <div className="w-full font-sans my-8 gap-4 overflow-x-auto whitespace-nowrap">
                    <h2 className="text-3xl">SAVED STREAMS</h2>
                    {user && user.vods.map((vod, idx) => (
                        <Link
                            key={vod}
                            className={`w-[300px] h-[180px] inline-block hover:cursor-pointer ${idx !== 0 ? "ml-4" : ""}`}
                            href={`/users/${params.userId}/vods/${vod}`}
                        >
                            <div className="flex flex-col items-center justify-center h-full bg-black bg-opacity-50 rounded-md">
                                <Image
                                    alt="vod icon"
                                    src={"/icons/video.svg"}
                                    width={100}
                                    height={100}
                                />
                                <p className="text-white">Streamed on {vod}</p>
                            </div>
                        </Link>
                    ))}
                </div>
            </div>
            <div className="w-full mx-4 h-screen bg-black bg-opacity-50"></div>
        </div>
    );
}
