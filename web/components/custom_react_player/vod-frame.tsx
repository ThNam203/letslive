export type { VideoInfo, VideoPlayerHandle } from "./video-player";

import { VideoPlayer, VideoInfo, VideoPlayerHandle } from "./video-player";
import { ClassValue } from "clsx";

export function VODFrame({
    videoInfo,
    className,
    onVideoStart,
    onProgressSeconds,
    enableSkipButtons,
    controlRef,
    startAt,
}: {
    videoInfo: VideoInfo;
    className?: ClassValue;
    onVideoStart?: () => void;
    onProgressSeconds?: (seconds: number) => void;
    enableSkipButtons?: boolean;
    controlRef?: React.Ref<VideoPlayerHandle>;
    startAt?: number;
}) {
    return (
        <VideoPlayer
            videoInfo={videoInfo}
            mode="vod"
            className={className}
            onVideoStart={onVideoStart}
            onProgressSeconds={onProgressSeconds}
            enableSkipButtons={enableSkipButtons}
            controlRef={controlRef}
            startAt={startAt}
        />
    );
}
