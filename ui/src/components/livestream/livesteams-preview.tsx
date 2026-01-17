"use client";

import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import LivestreamPreviewView from "./livestream-preview";
import { GetPopularLivestreams } from "../../lib/api/livestream";
import { Livestream } from "../../types/livestream";
import IconChevronDown from "../icons/chevron-down";
import { Separator } from "../ui/separator";
import IconPlay from "../icons/play";
import useT from "@/src/hooks/use-translation";
import { Card, CardContent } from "../ui/card";
import { Skeleton } from "../ui/skeleton";

const LivestreamsPreviewView = () => {
    const [isLoading, setIsLoading] = useState(false);
    const [limitView, setLimitView] = useState<number>(4);
    const [livestreams, setLivestreams] = useState<Livestream[]>([]);
    const { t } = useT(["common", "api-response", "fetch-error"]);

    useEffect(() => {
        const fetchLivestreams = async () => {
            await GetPopularLivestreams()
                .then((res) => {
                    if (res.success) {
                        setLivestreams(res.data ?? []);
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
        fetchLivestreams();
    }, []);

    if (isLoading) {
        return <LoadingSkeleton />;
    }

    return (
        <div className="flex flex-col gap-2 pr-2">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                {livestreams.slice(0, limitView).map((livestream, idx) => (
                    <LivestreamPreviewView key={idx} livestream={livestream} />
                ))}
            </div>
            {livestreams.length > limitView && (
                <StreamsSeparator
                    onClick={() => setLimitView((prev) => prev + 8)}
                />
            )}
            {livestreams.length == 0 && (
                <div className="flex flex-col items-center justify-center px-4 py-16 text-center">
                    <div className="mb-6 rounded-full bg-muted p-4">
                        <IconPlay className="text-muted-foreground h-16 w-16" />
                    </div>
                    <h2 className="mb-2 text-2xl font-semibold">
                        {t("common:no_livestreams")}
                    </h2>
                    <p className="text-muted-foreground max-w-md">
                        {t("common:no_livestreams_description")}
                    </p>
                </div>
            )}
        </div>
    );
};

const StreamsSeparator = ({ onClick }: { onClick: () => void }) => {
    const { t } = useT(["common"]);

    return (
        <div className="flex w-full flex-row items-center justify-between gap-4">
            <Separator className="flex-1" />
            <button
                className="flex flex-row items-center justify-center text-nowrap rounded-md border-border px-2 py-1 text-xs font-semibold text-primary duration-100 ease-linear hover:text-primary-hover"
                onClick={onClick}
            >
                <span className="">{t("show_more")}</span>
                <IconChevronDown />
            </button>
            <Separator className="flex-1" />
        </div>
    );
};

function LoadingSkeleton() {
    return (
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
    );
}

export default LivestreamsPreviewView;
