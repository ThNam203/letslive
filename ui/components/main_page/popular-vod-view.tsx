"use client";

import { useState, useEffect } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Clock, Eye, Film } from "lucide-react";
import { UserVOD } from "@/types/user";
import Image from "next/image";

export function PopularVODView() {
    const [isLoading, setIsLoading] = useState(false);
    const [vods, setVods] = useState<UserVOD[]>([]);

    if (isLoading) {
        return <VODSkeleton />;
    }

    if (vods.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center py-16 px-4 text-center">
                <div className="bg-muted/30 p-6 rounded-full mb-6">
                    <Film className="h-12 w-12 text-muted-foreground" />
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
            <h2 className="text-2xl font-semibold mb-6">Popular Videos</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                {vods.map((vod) => (
                    <VODCard key={vod.id} vod={vod} />
                ))}
            </div>
        </div>
    );
}

function VODCard({ vod }: { vod: any }) {
    return (
        <Card className="overflow-hidden transition-all hover:shadow-md">
            <div className="relative aspect-video bg-muted">
                <div className="absolute bottom-2 right-2">
                    <Badge
                        variant="secondary"
                        className="bg-black/70 text-white"
                    >
                        {vod.duration}
                    </Badge>
                </div>
                <Image
                    src={vod.thumbnail || "/placeholder.svg"}
                    alt={vod.title}
                    className="w-full h-full object-cover"
                />
            </div>
            <CardContent className="p-4">
                <div className="flex items-start gap-3">
                    <div className="h-10 w-10 rounded-full overflow-hidden bg-muted flex-shrink-0">
                        <Image
                            src={vod.avatar || "/placeholder.svg"}
                            alt={`${vod.channel} avatar`}
                            className="w-full h-full object-cover"
                        />
                    </div>
                    <div className="flex-1 min-w-0">
                        <h3 className="font-semibold text-base truncate">
                            {vod.title}
                        </h3>
                        <p className="text-sm text-muted-foreground truncate">
                            {vod.channel}
                        </p>
                        <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
                            <div className="flex items-center gap-1">
                                <Eye className="h-3 w-3" />
                                <span>{vod.views} views</span>
                            </div>
                            <div className="flex items-center gap-1">
                                <Clock className="h-3 w-3" />
                                <span>{vod.date}</span>
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
