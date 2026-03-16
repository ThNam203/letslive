"use client";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "@/components/utils/toast";
import { VideoInfo } from "@/components/custom_react_player/streaming-frame";
import { VODFrame } from "@/components/custom_react_player/vod-frame";
import VODCard from "@/components/livestream/vod-card";
import { VOD } from "@/types/vod";
import { PublicUser } from "@/types/user";
import { GetPublicVODsOfUser, GetVODInformation } from "@/lib/api/vod";
import { GetUserById } from "@/lib/api/user";
import ProfileView from "@/app/[lng]/(main)/users/[userId]/profile";
import useT from "@/hooks/use-translation";
import CommentSection from "@/components/vod-comments/comment-section";

export default function VODPage() {
    const { t } = useT(["fetch-error", "api-response"]);
    const [user, setUser] = useState<PublicUser | null>(null);
    const [vods, setVods] = useState<VOD[]>([]);
    const [isExtraOpen, setIsExtraOpen] = useState(false);

    const updateUser = (newUserInfo: PublicUser) => {
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
            await GetVODInformation(params.vodId)
                .then((res) => {
                    if (res.success) {
                        setPlayerInfo((prev) => ({
                            ...prev,
                            videoTitle: res.data?.title ?? "",
                            videoUrl: res.data?.playbackUrl ?? null,
                        }));
                    } else {
                        toast(t(`api-response:${res.key}`), {
                            toastId: res.requestId,
                            type: "error",
                        });
                    }
                })
                .catch((_) => {
                    toast(t("fetch-error:client_fetch_error"), {
                        toastId: "client-fetch-error-id",
                        type: "error",
                    });
                });
        };

        fetchVODInfo();
    }, [params.vodId]);

    useEffect(() => {
        if (!user) {
            return;
        }

        const fetchVODs = async () => {
            await GetPublicVODsOfUser(user.id)
                .then((res) => {
                    if (res.success) {
                        setVods(res.data ?? []);
                    } else {
                        toast(t(`api-response:${res.key}`), {
                            toastId: res.requestId,
                            type: "error",
                        });
                    }
                })
                .catch((_) => {
                    toast(t("fetch-error:client_fetch_error"), {
                        toastId: "client-fetch-error-id",
                        type: "error",
                    });
                });
        };

        fetchVODs();
    }, [user]);

    useEffect(() => {
        const fetchUserInfo = async () => {
            await GetUserById(params.userId)
                .then((userRes) => {
                    if (userRes.success) {
                        setUser(userRes.data ?? null);

                        setPlayerInfo((prev) => ({
                            ...prev,
                            streamer: {
                                name:
                                    userRes.data?.displayName ??
                                    userRes.data?.username ??
                                    "Streamer",
                            },
                        }));
                    } else {
                        toast(t(`api-response:${userRes.key}`), {
                            toastId: userRes.requestId,
                            type: "error",
                        });
                    }
                })
                .catch((_) => {
                    toast(t("fetch-error:client_fetch_error"), {
                        toastId: "client-fetch-error-id",
                        type: "error",
                    });
                });
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
                <CommentSection
                    vodId={params.vodId}
                    vodOwnerId={params.userId}
                    className="mt-4 pb-8"
                />
            </div>
            <div
                className={`bg-background fixed top-0 right-2 z-40 h-[100%-48px] w-full transition-all duration-300 md:relative md:w-80 lg:w-96 ${isExtraOpen ? "translate-x-0" : "translate-x-full md:translate-x-0"}`}
            >
                <div className="border-border bg-background flex h-full w-full flex-col border-x font-sans">
                    <h2 className="p-4 font-semibold">Other streams</h2>
                    <div className="small-scrollbar h-full overflow-y-auto px-4">
                        {vods
                            .filter((v) => v.id !== params.vodId)
                            .map((vod, idx) => (
                                <VODCard
                                    key={idx}
                                    vod={vod}
                                    variant="with-user"
                                    className="mb-2"
                                />
                            ))}
                    </div>
                </div>
            </div>
        </div>
    );
}
