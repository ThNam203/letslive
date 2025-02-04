"use client";
import {
    StreamingFrame,
    VideoInfo,
} from "@/components/custom_react_player/streaming_frame";
import { VODFrame } from "@/components/custom_react_player/vod_frame";
import { Button } from "@/components/ui/button";
import { GetUserById } from "@/lib/api/user";
import { cn } from "@/lib/utils";
import { User } from "@/types/user";
import Image from "next/image";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";

const servers = ["", "localhost:8890", "localhost:8891"];

export default function VODPage() {
    const [user, setUser] = useState<User | null>(null);
    const [serverIndex, setServerIndex] = useState(0);
    const params = useParams<{ userId: string; vodDate: string }>();
    const [playerInfo, setPlayerInfo] = useState<VideoInfo>({
        videoTitle: "Live Streaming",
        streamer: {
            name: "Dr. Pedophile",
        },
        videoUrl: null,
    });

    useEffect(() => {
        const newUrl =
            serverIndex == 0
                ? `http://localhost:8889/static/${params.userId}/vods/${params.vodDate}/index.m3u8`
                : `http://localhost:8889/static/${params.userId}/vods/${params.vodDate}/${servers[serverIndex]}_index.m3u8`;

        setPlayerInfo((prev) => ({
            ...prev,
            videoUrl: newUrl,
        }));
    }, [params.userId, params.vodDate]);

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

    useEffect(() => {
        const newUrl =
            serverIndex == 0
                ? `http://localhost:8889/static/${params.userId}/vods/${params.vodDate}/index.m3u8`
                : `http://localhost:8889/static/${params.userId}/vods/${params.vodDate}/${servers[serverIndex]}_index.m3u8`;

        setPlayerInfo((prev) => ({
            ...prev,
            videoUrl: newUrl,
        }));
    }, [serverIndex]);

    return (
        <div className="overflow-y-auto h-full flex lg:flex-row max-lg:flex-col">
            <div className="w-[1200px] min-w-[1200px]">
                <div className="w-full h-[675px] bg-black">
                    <VODFrame videoInfo={playerInfo} />
                </div>
                <div className="w-full font-sans mt-4 overflow-x-auto whitespace-nowrap">
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
                <div className="w-full font-sans my-8 gap-4 overflow-x-auto whitespace-nowrap">
                    <h2 className="text-3xl mb-4">OTHER SAVED STREAMS</h2>
                    {user &&
                        user.vods
                            .filter((v) => v !== params.vodDate)
                            .map((vod, idx) => (
                                <Link
                                    key={vod}
                                    className={`w-[300px] h-[180px] inline-block hover:cursor-pointer ${
                                        idx !== 0 ? "ml-4" : ""
                                    }`}
                                    href={`/users/${params.userId}/vods/${vod}`}
                                >
                                    <div className="flex flex-col items-center justify-center h-full bg-black bg-opacity-50 rounded-md">
                                        <Image
                                            alt="vod icon"
                                            src={"/icons/video.svg"}
                                            width={100}
                                            height={100}
                                        />
                                        <p className="text-white">
                                            Streamed on {vod}
                                        </p>
                                    </div>
                                </Link>
                            ))}
                </div>
            </div>
            <div className="w-full mx-4 h-screen bg-black bg-opacity-50"></div>
        </div>
    );
}
