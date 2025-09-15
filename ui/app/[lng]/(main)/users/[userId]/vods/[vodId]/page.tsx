"use client";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import { VideoInfo } from "@/components/custom_react_player/streaming-frame";
import { VODFrame } from "@/components/custom_react_player/vod-frame";
import VODView from "@/components/livestream/vod";
import { VOD } from "@/types/vod";
import { User } from "@/types/user";
import { GetPublicVODsOfUser, GetVODInformation } from "@/lib/api/vod";
import { GetUserById } from "@/lib/api/user";
import ProfileView from "@/app/[lng]/(main)/users/[userId]/profile";

export default function VODPage() {
    const [user, setUser] = useState<User | null>(null);
    const [vods, setVods] = useState<VOD[]>([]);
    const [isExtraOpen, setIsExtraOpen] = useState(false);

    const updateUser = (newUserInfo: User) => {
        setUser((prev) => {
            if (prev)
                return {
                    ...prev,
                    ...newUserInfo,
                };

            return newUserInfo;
        });
    };
    const params = useParams<{ userId: string; vodId: string }>();

    const [playerInfo, setPlayerInfo] = useState<VideoInfo>({
        videoTitle: "Live Streaming",
        streamer: {
            name: "Streamer",
        },
        videoUrl: null,
    });

    useEffect(() => {
        const fetchVODInfo = async () => {
            const { vod, fetchError } = await GetVODInformation(params.vodId);

            if (fetchError) {
                toast.error(fetchError.message, {
                    toastId: "vod-fetch-error",
                });
                return;
            }

            setPlayerInfo((prev) => ({
                ...prev,
                videoTitle: vod!.title,
                videoUrl: vod!.playbackUrl,
            }));
        };

        fetchVODInfo();
    }, [params.vodId]);

    useEffect(() => {
        if (!user) {
            return;
        }

        const fetchVODs = async () => {
            const { vods, fetchError } = await GetPublicVODsOfUser(user.id);

            if (fetchError) {
                toast(fetchError.message, {
                    toastId: "vod-fetch-error",
                    type: "error",
                });
            } else {
                setVods(vods);
            }
        };

        fetchVODs();
    }, [user]);

    useEffect(() => {
        const fetchUserInfo = async () => {
            const { user, fetchError } = await GetUserById(params.userId);
            if (fetchError != undefined) {
                toast.error(fetchError.message, {
                    toastId: "user-fetch-error",
                });
            } else {
                setUser(user!);
                setPlayerInfo((prev) => ({
                    ...prev,
                    streamer: {
                        name: user?.displayName ?? user?.username ?? "Streamer",
                    },
                }));
            }
        };

        fetchUserInfo();
    }, [params.userId]);

    return (
        <div className="ml-4 flex h-full gap-6 overflow-hidden">
            {/* Main content area */}
            <div className="no-scrollbar flex-1 overflow-auto">
                <VODFrame videoInfo={playerInfo} className="mt-1" />
                {user && (
                    <ProfileView
                        user={user}
                        updateUser={updateUser}
                        vods={vods.filter((v) => v.id !== params.vodId)}
                        showRecentActivity={false}
                        className="mt-2"
                    />
                )}
            </div>
            <div
                className={`fixed right-2 top-0 z-40 h-[100%-48px] w-full bg-background transition-all duration-300 md:relative md:w-80 lg:w-96 ${isExtraOpen ? "translate-x-0" : "translate-x-full md:translate-x-0"}`}
            >
                <div className="flex h-full w-full flex-col border-x border-border bg-background font-sans">
                    <h2 className="p-4 font-semibold">Other streams</h2>
                    <div className="small-scrollbar h-full overflow-y-auto px-4">
                        {vods.map((vod, idx) => (
                            <VODView key={idx} vod={vod} classname="mb-2" />
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
}
