"use client";
import { useParams } from "next/navigation";
import { useMemo, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { VideoInfo } from "@/components/custom_react_player/streaming-frame";
import { VODFrame } from "@/components/custom_react_player/vod-frame";
import MediaCard from "@/components/livestream/media-card";
import { VOD } from "@/types/vod";
import { PublicUser } from "@/types/user";
import { RegisterVODView } from "@/lib/api/vod";
import ProfileView from "@/app/[lng]/(main)/users/[userId]/profile";
import useT from "@/hooks/use-translation";
import CommentSection from "@/components/vod-comments/comment-section";
import { publicUserQueryKey, usePublicUser } from "@/hooks/queries/use-users";
import {
    publicVodsOfUserQueryKey,
    usePublicVodsOfUser,
    useVod,
} from "@/hooks/queries/use-vods";
import QueryError from "@/components/utils/query-error";

export default function VODPage() {
    const { t } = useT(["fetch-error", "api-response", "common"]);
    const params = useParams<{ userId: string; vodId: string }>();
    const queryClient = useQueryClient();
    const [isExtraOpen, setIsExtraOpen] = useState(false);
    // Keyed by VOD id rather than a plain boolean: navigating between VODs
    // reuses this component, and a boolean would carry the previous VOD's
    // "already counted" over to the next one.
    const [registeredVodId, setRegisteredVodId] = useState<string | null>(null);
    // a set, not one id: navigating away and back while a registration is
    // still in flight would otherwise let the same VOD be counted twice
    const registeringVodIdsRef = useRef(new Set<string>());

    const vodQuery = useVod(params.vodId);
    const userQuery = usePublicUser(params.userId);
    const vodsQuery = usePublicVodsOfUser(params.userId);

    const vod = vodQuery.data;
    const user = userQuery.data;
    const vods = vodsQuery.data;

    const vodDuration = vod?.duration ?? 0;

    const updateUser = (newUserInfo: PublicUser) => {
        queryClient.setQueryData<PublicUser>(
            publicUserQueryKey(params.userId),
            (prev) => (prev ? { ...prev, ...newUserInfo } : newUserInfo),
        );
    };

    const playerInfo: VideoInfo = useMemo(
        () => ({
            videoTitle: vod?.title ?? "",
            streamer: { name: user?.username ?? "" },
            videoUrl: vod?.playbackUrl ?? null,
        }),
        [vod?.title, vod?.playbackUrl, user?.username],
    );

    const otherVods = useMemo(
        () => (vods ?? []).filter((item) => item.id !== params.vodId),
        [vods, params.vodId],
    );

    const getViewThreshold = () => {
        let threshold = 15;
        const tenPercent = Math.floor(vodDuration * 0.1);
        if (tenPercent < threshold) {
            threshold = tenPercent;
        }
        if (vodDuration > 0 && threshold < 1) {
            threshold = 1;
        }
        return threshold;
    };

    const handleVODProgress = async (playedSeconds: number) => {
        if (
            !params.vodId ||
            registeredVodId === params.vodId ||
            registeringVodIdsRef.current.has(params.vodId)
        ) {
            return;
        }

        const threshold = getViewThreshold();
        if (Math.floor(playedSeconds) < threshold) {
            return;
        }

        const watchedSeconds = Math.floor(playedSeconds);
        registeringVodIdsRef.current.add(params.vodId);
        const res = await RegisterVODView(params.vodId, watchedSeconds).catch(
            () => null,
        );

        if (res?.success) {
            setRegisteredVodId(params.vodId);
            queryClient.setQueryData<VOD[]>(
                publicVodsOfUserQueryKey(params.userId),
                (prev) =>
                    prev?.map((item) =>
                        item.id === params.vodId
                            ? { ...item, viewCount: item.viewCount + 1 }
                            : item,
                    ),
            );
            return;
        }

        registeringVodIdsRef.current.delete(params.vodId);
    };

    return (
        <div className="ml-4 flex h-full gap-6 overflow-hidden">
            {/* Main content area */}
            <div className="no-scrollbar flex-1 overflow-auto">
                {vodQuery.isError ? (
                    // a blank player reads as a broken video rather than a
                    // request that never landed
                    <QueryError
                        className="mt-1"
                        onRetry={() => vodQuery.refetch()}
                    />
                ) : (
                    <VODFrame
                        videoInfo={playerInfo}
                        className="mt-1"
                        onProgressSeconds={handleVODProgress}
                    />
                )}

                {userQuery.isError && (
                    <QueryError
                        className="mt-2"
                        onRetry={() => userQuery.refetch()}
                    />
                )}

                {user && (
                    <ProfileView
                        user={user}
                        updateUser={updateUser}
                        vods={otherVods}
                        showRecentActivity={false}
                        className="mt-2"
                    />
                )}
                <CommentSection
                    key={params.vodId}
                    vodId={params.vodId}
                    vodOwnerId={params.userId}
                    className="mt-4 pb-8"
                />
            </div>
            <div
                className={`bg-background fixed top-0 right-2 z-40 h-[100%-48px] w-full transition-all duration-300 md:relative md:w-80 lg:w-96 ${isExtraOpen ? "translate-x-0" : "translate-x-full md:translate-x-0"}`}
            >
                <div className="border-border bg-background flex h-full w-full flex-col border-x font-sans">
                    <h2 className="p-4 font-semibold">
                        {t("common:other_streams")}
                    </h2>
                    <div className="small-scrollbar h-full overflow-y-auto px-4">
                        {vodsQuery.isError && (
                            <QueryError onRetry={() => vodsQuery.refetch()} />
                        )}
                        {otherVods.map((item) => (
                            <MediaCard
                                key={item.id}
                                kind="vod"
                                vod={item}
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
