"use client";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import { User } from "../../../../../../types/user";
import { VideoInfo } from "../../../../../../components/custom_react_player/streaming-frame";
import { GetAllLivestreamOfUser, GetVODInformation } from "../../../../../../lib/api/livestream";
import { GetUserById } from "../../../../../../lib/api/user";
import { VODFrame } from "../../../../../../components/custom_react_player/vod-frame";
import ProfileView from "../../profile";
import VODLink from "../../../../../../components/livestream/vod";
import GLOBAL from "../../../../../../global";
import { Livestream } from "../../../../../../types/livestream";

export default function VODPage() {
    const [user, setUser] = useState<User | null>(null);
    const [vods, setVods] = useState<Livestream[]>([]);
    const [isExtraOpen, setIsExtraOpen] = useState(false)

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
            const { vod, fetchError } = await GetVODInformation(
                params.vodId
            );

            if (fetchError) {
                toast.error(fetchError.message, {
                    toastId: "vod-fetch-error",
                });
                return;
            }

            const newUrl = `${GLOBAL.API_URL}/transcode/vods/${params.vodId}/index.m3u8`;
            setPlayerInfo((prev) => ({
                ...prev,
                videoTitle: vod!.title,
                videoUrl: newUrl,
            }));
        };

        fetchVODInfo();
    }, [params.vodId]);

    useEffect(() => {
        if (!user) {
            return;
        }

        const fetchVODs = async () => {
            const { livestreams, fetchError } = await GetAllLivestreamOfUser(user.id);

            if (fetchError) {
                toast(fetchError.message, {
                    toastId: "vod-fetch-error",
                    type: "error",
                });
            } else {
                setVods(livestreams);
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
                    }
                }))
            }
        };

        fetchUserInfo();
    }, [params.userId]);

    return (

        <div className="flex h-full overflow-hidden ml-4 gap-6">
            {/* Main content area */}
            <div className="flex-1 overflow-auto no-scrollbar">
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
                className={`w-full h-[100%-48px] md:w-80 lg:w-96 bg-background transition-all duration-300 fixed md:relative top-0 right-2 z-40 ${isExtraOpen ? "translate-x-0" : "translate-x-full md:translate-x-0"}`}
            >
                <div className="w-full h-full flex flex-col font-sans border-x bg-background border-border">
                    <h2 className="font-semibold p-4">
                        Other streams
                    </h2>
                    <div className="overflow-y-auto h-full px-4 small-scrollbar">
                        {vods
                            ?.filter(
                                (v) =>
                                    v.id !== params.vodId &&
                                    v.status !== "live"
                            )
                            .map((vod, idx) => (
                                <VODLink key={idx} vod={vod} classname="mb-2" />
                            ))}
                    </div>
                </div>
            </div>
        </div>
    );
}