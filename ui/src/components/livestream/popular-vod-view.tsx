"use client";

import { useState, useEffect } from "react";
import { Card, CardContent } from "../ui/card";
import { Skeleton } from "../ui/skeleton";
import { GetPopularVODs } from "../../lib/api/vod";
import { toast } from "react-toastify";
import IconFilm from "../icons/film";
import { VOD } from "@/src/types/vod";
import useT from "@/src/hooks/use-translation";
import VODCard from "./vod-card";

export function PopularVODView() {
    const [isLoading, setIsLoading] = useState(false);
    const [vods, setVods] = useState<VOD[]>([]);
    const { t } = useT(["common", "api-response", "fetch-error"]);

    useEffect(() => {
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
                    toast(t("fetch-error:client_fetch_error"), {
                        toastId: "client-fetch-error-id",
                        type: "error",
                    });
                })
                .finally(() => {
                    setIsLoading(false);
                });
        };

        setIsLoading(true);
        fetchData();
    }, []);

    if (isLoading) {
        return <LoadingSkeleton />;
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
                    <VODCard key={vod.id} vod={vod} variant="with-user" />
                ))}
            </div>
        </div>
    );
}

function LoadingSkeleton() {
    return (
        <div>
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                {[1, 2, 3, 4].map((i) => (
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
