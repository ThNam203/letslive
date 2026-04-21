export { formatTime } from "./video-player";
export type { VideoInfo } from "./video-player";

import { VideoPlayer, VideoInfo } from "./video-player";
import { ClassValue } from "clsx";

export function StreamingFrame({
    videoInfo,
    className,
    onVideoStart,
    enableSkipButtons,
}: {
    videoInfo: VideoInfo;
    className?: ClassValue;
    onVideoStart?: () => void;
    enableSkipButtons?: boolean;
}) {
    return (
        <VideoPlayer
            videoInfo={videoInfo}
            mode="live"
            className={className}
            onVideoStart={onVideoStart}
            enableSkipButtons={enableSkipButtons}
        />
    );
}
