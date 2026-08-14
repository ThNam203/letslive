"use client";

import { Card, CardContent } from "../ui/card";
import { Skeleton } from "../ui/skeleton";
import { Button } from "../ui/button";
import IconFilm from "../icons/film";
import useT from "@/hooks/use-translation";
import MediaCard from "./media-card";
import { useVodsInfinite } from "@/hooks/queries/use-vods-infinite";

export function VodFeedView() {
    const { data, isLoading, isFetchingNextPage, hasNextPage, fetchNextPage } =
        useVodsInfinite();
    const vods = data?.pages.flat() ?? [];
    const { t } = useT(["common"]);

    if (isLoading) {
        return <LoadingSkeleton />;
    }

    if (vods.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center px-4 py-16 text-center">
                <div className="bg-muted mb-6 rounded-full p-6">
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
                    <MediaCard
                        key={vod.id}
                        kind="vod"
                        vod={vod}
                        variant="with-user"
                    />
                ))}
            </div>
            {hasNextPage && (
                <div className="mt-4 flex justify-center">
                    <Button
                        variant="ghost"
                        onClick={() => fetchNextPage()}
                        disabled={isFetchingNextPage}
                    >
                        {isFetchingNextPage
                            ? t("common:loading")
                            : t("common:show_more")}
                    </Button>
                </div>
            )}
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
