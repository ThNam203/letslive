"use client";

import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import LivestreamPreviewView from "./livestream-preview";
import { GetPopularLivestreams } from "../../lib/api/livestream";
import { Livestream } from "../../types/livestream";
import IconChevronDown from "../icons/chevron-down";
import { Separator } from "../ui/separator";
import IconPlay from "../icons/play";

const LivestreamsPreviewView = () => {
    const [limitView, setLimitView] = useState<number>(4);
    const [livestreams, setLivestreams] = useState<Livestream[]>([]);

    useEffect(() => {
        const fetchLivestreams = async () => {
            const { livestreams, fetchError } = await GetPopularLivestreams();
            if (fetchError) {
                toast(fetchError.message, {
                    toastId: "livestreams-fetch-error",
                    type: "error",
                });
            }

            setLivestreams(livestreams ?? []);
        };

        fetchLivestreams();
    }, []);

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
                        No Live Streams
                    </h2>
                    <p className="text-muted-foreground max-w-md">
                        There is currently no one streaming. Check back later or
                        explore our video on demand content.
                    </p>
                </div>
            )}
        </div>
    );
};

const StreamsSeparator = ({ onClick }: { onClick: () => void }) => {
    return (
        <div className="flex w-full flex-row items-center justify-between gap-4">
            <Separator className="flex-1" />
            <button
                className="flex flex-row items-center justify-center text-nowrap rounded-md border-border px-2 py-1 text-xs font-semibold text-primary duration-100 ease-linear hover:text-primary-hover"
                onClick={onClick}
            >
                <span className="">Show more</span>
                <IconChevronDown />
            </button>
            <Separator className="flex-1" />
        </div>
    );
};

export default LivestreamsPreviewView;
