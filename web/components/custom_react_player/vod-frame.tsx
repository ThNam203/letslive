export type { VideoInfo } from "./video-player";

import { VideoPlayer, VideoInfo } from "./video-player";
import { ClassValue } from "clsx";

export function VODFrame({
    videoInfo,
    className,
    onVideoStart,
    onProgressSeconds,
    enableSkipButtons,
}: {
    videoInfo: VideoInfo;
    className?: ClassValue;
    onVideoStart?: () => void;
    onProgressSeconds?: (seconds: number) => void;
    enableSkipButtons?: boolean;
}) {
    return (
        <VideoPlayer
            videoInfo={videoInfo}
            mode="vod"
            className={className}
            onVideoStart={onVideoStart}
            onProgressSeconds={onProgressSeconds}
            enableSkipButtons={enableSkipButtons}
        />
    );
}
