"use client";

import { useState, useEffect } from "react";
import { Card, CardContent } from "../ui/card";
import { Badge } from "../ui/badge";
import { Skeleton } from "../ui/skeleton";
import { GetPopularVODs } from "../../lib/api/vod";
import { toast } from "react-toastify";
import { User } from "../../types/user";
import { GetUserById } from "../../lib/api/user";
import { dateDiffFromNow, formatSeconds } from "@/utils/timeFormats";
import { useParams, useRouter } from "next/navigation";
import GLOBAL from "../../global";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import IconFilm from "../icons/film";
import IconEye from "../icons/eye";
import IconClock from "../icons/clock";
import LiveImage from "./live-image";
import { VOD } from "@/types/vod";
import useT from "@/hooks/use-translation";

export function PopularVODView() {
    const [isLoading, setIsLoading] = useState(false);
    const [vods, setVods] = useState<VOD[]>([]);
    const { t } = useT(["common", "api-response", "fetch-error"]);

    useEffect(() => {
        setIsLoading(true);
        const fetchData = async () => {
            await GetPopularVODs()
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
                    toast(t("fetch-error:client_fetch_error", { type: "error" }));
                })
                .finally(() => {
                    setIsLoading(false);
                });
        };

        fetchData();
    }, []);

    if (isLoading) {
        return <VODSkeleton />;
    }

    if (vods.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center px-4 py-16 text-center">
                <div className="mb-6 rounded-full bg-muted p-6">
                    <IconFilm className="text-muted-foreground h-12 w-12" />
                </div>
                <h2 className="mb-2 text-2xl font-semibold">
                    {t("common:no_videos")}
                </h2>
                <p className="text-muted-foreground max-w-md">
                    {t("common:no_videos_description")}
                </p>
            </div>
        );
    }

    return (
        <div>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                {vods.map((vod) => (
                    <VODCard key={vod.id} vod={vod} />
                ))}
            </div>
        </div>
    );
}

function VODCard({ vod }: { vod: VOD }) {
    const [user, setUser] = useState<User | null>(null);
    const router = useRouter();

    useEffect(() => {
        const fetchUser = async () => {
            const res = await GetUserById(vod.userId);
            if (res.success) setUser(res.data ?? null);
        };

        fetchUser();
    }, [vod.userId]);

    return (
        <Card className="w-full overflow-hidden rounded-sm border-border transition-all hover:shadow-md">
            <div className="relative aspect-video bg-muted">
                <div className="absolute bottom-2 right-2">
                    <Badge
                        variant="secondary"
                        className="bg-black/70 text-white"
                    >
                        {formatSeconds(vod.duration)}
                    </Badge>
                </div>
                <LiveImage
                    src={
                        vod.thumbnailUrl ??
                        `${GLOBAL.API_URL}/files/livestreams/${vod.id}/thumbnail.jpeg`
                    }
                    alt={vod.title}
                    className="h-full w-full hover:cursor-pointer"
                    width={500}
                    height={500}
                    onClick={() =>
                        router.push(`/users/${vod.userId}/vods/${vod.id}`)
                    }
                    fallbackSrc="/images/streaming.jpg"
                    alwaysRefresh={false}
                />
            </div>
            <CardContent className="p-4">
                <div className="flex items-start gap-3">
                    <div className="h-10 w-10 flex-shrink-0 overflow-hidden rounded-full bg-muted">
                        <Avatar>
                            <AvatarImage
                                src={user?.profilePicture}
                                alt={`${user?.username} avatar`}
                                className="h-full w-full object-cover"
                                width={40}
                                height={40}
                            />
                            <AvatarFallback>
                                {user?.username.charAt(0).toUpperCase()}
                            </AvatarFallback>
                        </Avatar>
                    </div>
                    <div className="min-w-0 flex-1">
                        <h3 className="truncate text-base font-semibold">
                            {vod.title}
                        </h3>
                        <p className="text-muted-foreground truncate text-sm">
                            {user
                                ? (user.displayName ?? user.username)
                                : "Unknown"}
                        </p>
                        <div className="text-muted-foreground mt-1 flex items-center gap-3 text-xs">
                            <div className="flex items-center gap-1">
                                <IconEye className="h-3 w-3" />
                                <span>
                                    {vod.viewCount}{" "}
                                    {vod.viewCount < 2 ? "view" : "views"}
                                </span>
                            </div>
                            <div className="flex items-center gap-1">
                                <IconClock className="h-3 w-3" />
                                <span>
                                    {dateDiffFromNow(vod.createdAt)} ago
                                </span>
                            </div>
                        </div>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}

function VODSkeleton() {
    return (
        <div>
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                {[1, 2, 3, 4, 5, 6].map((i) => (
                    <Card key={i} className="overflow-hidden">
                        <Skeleton className="aspect-video w-full" />
                        <CardContent className="p-4">
                            <div className="flex items-start gap-3">
                                <Skeleton className="h-10 w-10 flex-shrink-0 rounded-full" />
                                <div className="flex-1">
                                    <Skeleton className="mb-2 h-5 w-full" />
                                    <Skeleton className="mb-2 h-4 w-3/4" />
                                    <Skeleton className="h-3 w-1/2" />
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                ))}
            </div>
        </div>
    );
}
