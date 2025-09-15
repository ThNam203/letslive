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
import { useRouter } from "next/navigation";
import GLOBAL from "../../global";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import IconFilm from "../icons/film";
import IconEye from "../icons/eye";
import IconClock from "../icons/clock";
import LiveImage from "./live-image";
import { VOD } from "@/types/vod";

export function PopularVODView() {
    const [isLoading, setIsLoading] = useState(false);
    const [vods, setVods] = useState<VOD[]>([]);

    useEffect(() => {
        setIsLoading(true);
        const fetchData = async () => {
            const { vods, fetchError } = await GetPopularVODs();
            if (fetchError) {
                toast("Failed to fetch popular videos", { type: "error" });
                return;
            } else {
                setVods(vods);
            }
            setIsLoading(false);
        };
        fetchData();
    }, []);

    if (isLoading) {
        return <VODSkeleton />;
    }

    if (vods.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center py-16 px-4 text-center">
                <div className="bg-muted p-6 rounded-full mb-6">
                    <IconFilm className="h-12 w-12 text-muted-foreground" />
                </div>
                <h2 className="text-2xl font-semibold mb-2">
                    No Videos Available
                </h2>
                <p className="text-muted-foreground max-w-md">
                    There are currently no videos available. Check back later
                    for new content.
                </p>
            </div>
        );
    }

    return (
        <div>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
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
            const { user } = await GetUserById(vod.userId);
            if (user) setUser(user);
        };

        fetchUser();
    }, [vod.userId]);

    return (
        <Card className="w-full overflow-hidden transition-all hover:shadow-md rounded-sm border-border">
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
                    src={vod.thumbnailUrl ?? `${GLOBAL.API_URL}/files/livestreams/${vod.id}/thumbnail.jpeg`}
                    alt={vod.title}
                    className="w-full h-full hover:cursor-pointer"
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
                    <div className="h-10 w-10 rounded-full overflow-hidden bg-muted flex-shrink-0">
                        <Avatar>
                            <AvatarImage
                                src={
                                    user?.profilePicture
                                }
                                alt={`${user?.username} avatar`}
                                className="w-full h-full object-cover"
                                width={40}
                                height={40}
                            />
                            <AvatarFallback>
                                {user?.username.charAt(0).toUpperCase()}
                            </AvatarFallback>
                        </Avatar>
                    </div>
                    <div className="flex-1 min-w-0">
                        <h3 className="font-semibold text-base truncate">
                            {vod.title}
                        </h3>
                        <p className="text-sm text-muted-foreground truncate">
                            {user
                                ? user.displayName ?? user.username
                                : "Unknown"}
                        </p>
                        <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
                            <div className="flex items-center gap-1">
                                <IconEye className="h-3 w-3" />
                                <span>{vod.viewCount} {vod.viewCount < 2 ? "view" : "views"}</span>
                            </div>
                            <div className="flex items-center gap-1">
                                <IconClock className="h-3 w-3" />
                                <span>{dateDiffFromNow(vod.createdAt)} ago</span>
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
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                {[1, 2, 3, 4, 5, 6].map((i) => (
                    <Card key={i} className="overflow-hidden">
                        <Skeleton className="aspect-video w-full" />
                        <CardContent className="p-4">
                            <div className="flex items-start gap-3">
                                <Skeleton className="h-10 w-10 rounded-full flex-shrink-0" />
                                <div className="flex-1">
                                    <Skeleton className="h-5 w-full mb-2" />
                                    <Skeleton className="h-4 w-3/4 mb-2" />
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
