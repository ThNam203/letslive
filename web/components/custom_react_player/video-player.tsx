"use client";
import { Slider } from "@/components/ui/slider";
import { ClassValue } from "clsx";
import { useRef, useState } from "react";
import ReactPlayer from "react-player";
import screenfull from "screenfull";
import { cn } from "@/utils/cn";
import {
    getResolutionHeight,
    formatResolutionForDisplay,
} from "@/utils/resolution";
import dynamic from "next/dynamic";
import useT from "@/hooks/use-translation";
import IconCheck from "../icons/check";
import IconFullscreen from "../icons/fullscreen";
import IconFullscreenExit from "../icons/fullscreen-exit";
import IconPause from "../icons/pause";
import IconPlay from "../icons/play";
import IconVolumeDown from "../icons/volume-down";
import IconVolumeOff from "../icons/volume-off";
import IconVolumeUp from "../icons/volume-up";
import IconFastForward from "../icons/fast-forward";
import IconLoader from "../icons/loader";

const ReactPlayerWrapper = dynamic(() => import("./react-player-wrapper"), {
    ssr: false,
});

export const formatTime = (seconds: number) => {
    if (isNaN(seconds) || seconds < 0) return "00:00";
    const hrs = Math.floor(seconds / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);
    if (hrs > 0) {
        return `${hrs.toString().padStart(2, "0")}:${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
    }
    return `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
};

const PLAYBACK_RATES: Record<string, number> = {
    "0.5x": 0.5,
    "1x": 1.0,
    "1.5x": 1.5,
    "2x": 2.0,
};

export type VideoInfo = {
    videoUrl: string | null;
    videoTitle: string;
    streamer: { name: string };
};

export type PlayerMode = "live" | "vod";

type PlayerConfig = {
    playbackRate: number;
    resolution: string;
    volumeValue: number;
    isFullscreen: boolean;
    loop: boolean;
};

// ─── Public component ────────────────────────────────────────────────────────

export function VideoPlayer({
    videoInfo,
    mode,
    className,
    onVideoStart,
    enableSkipButtons,
}: {
    videoInfo: VideoInfo;
    mode: PlayerMode;
    className?: ClassValue;
    onVideoStart?: () => void;
    enableSkipButtons?: boolean;
}) {
    const skipButtons = enableSkipButtons ?? mode === "vod";
    const containerRef = useRef<HTMLDivElement>(null);
    const playerRef = useRef<ReactPlayer>(null);

    const [idleCount, setIdleCount] = useState(0);
    const [duration, setDuration] = useState(0);
    const [isPlaying, setIsPlaying] = useState(false);
    const [isLoading, setIsLoading] = useState(true);
    const [currentTime, setCurrentTime] = useState(0);
    const [config, setConfig] = useState<PlayerConfig>({
        playbackRate: 1.0,
        resolution: "Auto",
        volumeValue: 100,
        isFullscreen: false,
        loop: false,
    });
    const [resolutions, setResolutions] = useState<string[]>(["Auto"]);

    const playVideo = () => {
        setIsPlaying(true);
        if (onVideoStart && currentTime === 0) onVideoStart();
    };

    const pauseVideo = () => setIsPlaying(false);

    const seekToTime = (time: number) => {
        if (playerRef.current) {
            setCurrentTime(time);
            playerRef.current.seekTo(time);
        }
    };

    const enterFullscreen = () => {
        if (screenfull.isEnabled && containerRef.current) {
            screenfull.request(containerRef.current);
            setConfig((prev) => ({ ...prev, isFullscreen: true }));
        }
    };

    const exitFullscreen = () => {
        if (screenfull.isEnabled) {
            screenfull.exit();
            setConfig((prev) => ({ ...prev, isFullscreen: false }));
        }
    };

    const handleResolutionChange = (value: string) => {
        if (!playerRef.current?.getInternalPlayer("hls")) return;
        setConfig((prev) => ({ ...prev, resolution: value }));
        if (value === "Auto") {
            playerRef.current.getInternalPlayer("hls").currentLevel = -1;
        } else {
            const selectedHeight = getResolutionHeight(value);
            if (selectedHeight === null) return;
            const levelIndex = resolutions.findIndex((reso) => {
                if (reso === "Auto") return false;
                return getResolutionHeight(reso) === selectedHeight;
            });
            if (levelIndex !== -1) {
                playerRef.current.getInternalPlayer("hls").currentLevel =
                    levelIndex - 1;
            }
        }
    };

    return (
        <div
            ref={containerRef}
            className={cn("relative aspect-video w-full bg-black", className)}
            onMouseMove={() => setIdleCount(0)}
            onClick={() => setIdleCount(0)}
        >
            {videoInfo.videoUrl != null && (
                <>
                    <ReactPlayerWrapper
                        playerRef={playerRef}
                        url={videoInfo.videoUrl}
                        muted={config.volumeValue === 0}
                        volume={config.volumeValue / 100}
                        playing={isPlaying}
                        width="100%"
                        height="100%"
                        playbackRate={config.playbackRate}
                        loop={config.loop}
                        onProgress={(state: any) => {
                            setIdleCount((c) => c + 1);
                            setCurrentTime(state.playedSeconds);
                        }}
                        onDuration={(d: number) => setDuration(d)}
                        onBuffer={() => setIsLoading(true)}
                        onBufferEnd={() => setIsLoading(false)}
                        config={{ file: { forceHLS: true } }}
                        onReady={(reactPlayer) => {
                            setIsLoading(false);
                            const hlsPlayer =
                                reactPlayer.getInternalPlayer("hls");
                            if (!hlsPlayer) return;
                            const newResolutions = hlsPlayer.levels.map(
                                (level: any) => level.attrs.RESOLUTION,
                            );
                            setResolutions(["Auto", ...newResolutions]);
                        }}
                    />
                    <PlayerOverlay
                        mode={mode}
                        isPlaying={isPlaying}
                        isLoading={isLoading}
                        currentTime={currentTime}
                        duration={duration}
                        config={config}
                        resolutions={resolutions}
                        enableSkipButtons={skipButtons}
                        videoInfo={videoInfo}
                        className={idleCount > 3 ? "opacity-0" : "opacity-100"}
                        onPlay={playVideo}
                        onPause={pauseVideo}
                        onSeek={seekToTime}
                        onVolumeChange={(v) =>
                            setConfig((prev) => ({ ...prev, volumeValue: v }))
                        }
                        onFullScreen={enterFullscreen}
                        onExitFullScreen={exitFullscreen}
                        onPlaybackRateChange={(v) =>
                            setConfig((prev) => ({ ...prev, playbackRate: v }))
                        }
                        onResolutionChange={handleResolutionChange}
                    />
                </>
            )}
        </div>
    );
}

// ─── Internal sub-components ─────────────────────────────────────────────────

type OverlayProps = {
    mode: PlayerMode;
    isPlaying: boolean;
    isLoading: boolean;
    currentTime: number;
    duration: number;
    config: PlayerConfig;
    resolutions: string[];
    enableSkipButtons: boolean;
    videoInfo: VideoInfo;
    className?: ClassValue;
    onPlay: () => void;
    onPause: () => void;
    onSeek: (time: number) => void;
    onVolumeChange: (value: number) => void;
    onFullScreen: () => void;
    onExitFullScreen: () => void;
    onPlaybackRateChange: (value: number) => void;
    onResolutionChange: (value: string) => void;
};

function PlayerOverlay(props: OverlayProps) {
    const { videoInfo, isPlaying, isLoading, onPlay, onPause, className } =
        props;
    return (
        <div
            className={cn(
                "border-border absolute inset-0 flex flex-col border transition-opacity duration-300",
                className,
            )}
        >
            <PlayerHeader videoInfo={videoInfo} />
            <PlayerCenter
                isPlaying={isPlaying}
                isLoading={isLoading}
                onPlay={onPlay}
                onPause={onPause}
            />
            <PlayerControls {...props} />
        </div>
    );
}

function PlayerHeader({ videoInfo }: { videoInfo: VideoInfo }) {
    return (
        <div className="absolute top-3 right-3 left-3 flex items-center justify-between gap-2 font-sans font-bold text-white sm:top-4 sm:right-4 sm:left-4">
            <span className="max-w-[55%] truncate rounded bg-black/70 px-2 py-1 text-xs sm:text-sm">
                {videoInfo.videoTitle}
            </span>
            <span className="max-w-[40%] truncate rounded bg-black/70 px-2 py-1 text-xs sm:text-sm">
                {videoInfo.streamer.name}
            </span>
        </div>
    );
}

function PlayerCenter({
    isPlaying,
    isLoading,
    onPlay,
    onPause,
}: {
    isPlaying: boolean;
    isLoading: boolean;
    onPlay: () => void;
    onPause: () => void;
}) {
    return (
        <div
            className="flex flex-1 items-center justify-center"
            onMouseUp={() => (isPlaying ? onPause() : onPlay())}
        >
            {isLoading ? (
                <div className="flex h-16 w-16 items-center justify-center rounded-full bg-black/30 sm:h-28 sm:w-28">
                    <IconLoader
                        width="60%"
                        height="60%"
                        className="cursor-pointer"
                        color="white"
                    />
                </div>
            ) : (
                !isPlaying && (
                    <div className="flex h-16 w-16 items-center justify-center rounded-full bg-black/30 sm:h-28 sm:w-28">
                        <IconPlay
                            width="60%"
                            height="60%"
                            className="cursor-pointer"
                            color="white"
                        />
                    </div>
                )
            )}
        </div>
    );
}

function PlayerControls({
    mode,
    isPlaying,
    currentTime,
    duration,
    config,
    resolutions,
    enableSkipButtons,
    onPlay,
    onPause,
    onSeek,
    onVolumeChange,
    onFullScreen,
    onExitFullScreen,
    onPlaybackRateChange,
    onResolutionChange,
}: OverlayProps) {
    const { t } = useT(["common", "accessibility"]);
    const isAtLive = mode === "live" && duration > 0 && duration - currentTime <= 3;

    return (
        <div className="flex w-full flex-col px-3 pb-3 sm:px-4 sm:pb-4">
            {/* Live badge — live mode only */}
            {mode === "live" && (
                <div
                    className={cn(
                        "mb-2 flex w-fit items-center gap-2 rounded bg-black/30 px-2 py-1 text-xs font-semibold text-red-500 sm:text-sm",
                        !isAtLive &&
                            "cursor-pointer transition-opacity hover:opacity-80",
                    )}
                    onClick={() => !isAtLive && onSeek(duration)}
                >
                    <div
                        className={cn(
                            "h-2 w-2 rounded-full",
                            isAtLive ? "bg-red-500" : "bg-gray-500",
                        )}
                    />
                    <span>{t("common:live")}</span>
                </div>
            )}

            {/* Progress bar */}
            <div className="mb-2 w-full cursor-pointer sm:mb-3">
                <Slider
                    value={[duration !== 0 ? (currentTime / duration) * 100 : 0]}
                    onValueChange={(value) =>
                        onSeek((value[0] / 100) * duration)
                    }
                    max={100}
                    step={0.1}
                    trackClassName="bg-gray-500/50"
                    rangeClassName="bg-white"
                    thumbClassName="border-white/50 bg-white"
                />
            </div>

            {/* Button row */}
            <div className="flex w-full items-center justify-between text-white">
                {/* Left group: play, volume, time */}
                <div className="flex items-center gap-2 sm:gap-4">
                    <ControlButton
                        onClick={() => (isPlaying ? onPause() : onPlay())}
                    >
                        {isPlaying ? (
                            <IconPause
                                width="20px"
                                height="20px"
                                color="white"
                            />
                        ) : (
                            <IconPlay
                                width="20px"
                                height="20px"
                                color="white"
                            />
                        )}
                    </ControlButton>

                    <VolumeControl onVolumeChange={onVolumeChange} />

                    {/* Time display — VOD only, hidden on small screens */}
                    {mode === "vod" && (
                        <span className="hidden h-8 items-center rounded bg-black/30 px-2 text-xs text-white sm:flex sm:h-10 sm:px-3 sm:text-sm">
                            {formatTime(currentTime)} / {formatTime(duration)}
                        </span>
                    )}
                </div>

                {/* Right group: skip, speed, resolution, fullscreen */}
                <div className="flex items-center gap-1 sm:gap-2">
                    {enableSkipButtons && (
                        <>
                            <ControlButton
                                onClick={() =>
                                    onSeek(Math.max(0, currentTime - 10))
                                }
                                title={t("accessibility:seek_back_10")}
                            >
                                <IconFastForward
                                    width="20px"
                                    height="20px"
                                    color="white"
                                    className="scale-x-[-1]"
                                />
                            </ControlButton>
                            <ControlButton
                                onClick={() =>
                                    onSeek(Math.min(duration, currentTime + 10))
                                }
                                title={t("accessibility:seek_forward_10")}
                            >
                                <IconFastForward
                                    width="20px"
                                    height="20px"
                                    color="white"
                                />
                            </ControlButton>
                        </>
                    )}

                    {/* Playback speed — VOD only */}
                    {mode === "vod" && (
                        <Combobox
                            options={Object.keys(PLAYBACK_RATES)}
                            value={config.playbackRate + "x"}
                            onChange={(v) =>
                                onPlaybackRateChange(
                                    PLAYBACK_RATES[
                                        v as keyof typeof PLAYBACK_RATES
                                    ],
                                )
                            }
                        />
                    )}

                    <Combobox
                        options={resolutions
                            .map((r) => formatResolutionForDisplay(r))
                            .filter((r): r is string => r !== null)}
                        value={
                            formatResolutionForDisplay(config.resolution) ||
                            "Auto"
                        }
                        onChange={onResolutionChange}
                    />

                    <ControlButton
                        onClick={() =>
                            screenfull.isFullscreen
                                ? onExitFullScreen()
                                : onFullScreen()
                        }
                    >
                        {config.isFullscreen ? (
                            <IconFullscreenExit
                                width="20px"
                                height="20px"
                                color="white"
                            />
                        ) : (
                            <IconFullscreen
                                width="20px"
                                height="20px"
                                color="white"
                            />
                        )}
                    </ControlButton>
                </div>
            </div>
        </div>
    );
}

function ControlButton({
    onClick,
    title,
    children,
}: {
    onClick: () => void;
    title?: string;
    children: React.ReactNode;
}) {
    return (
        <div
            onClick={onClick}
            title={title}
            className="flex h-8 w-8 cursor-pointer items-center justify-center rounded bg-black/30 transition-colors hover:bg-black/50 sm:h-10 sm:w-10"
        >
            {children}
        </div>
    );
}

function VolumeControl({
    onVolumeChange,
}: {
    onVolumeChange: (value: number) => void;
}) {
    const [volume, setVolume] = useState(100);
    const lastNonZeroRef = useRef(100);

    const handleChange = (value: number) => {
        setVolume(value);
        onVolumeChange(value);
        if (value !== 0) lastNonZeroRef.current = value;
    };

    const VolumeIcon =
        volume === 0
            ? IconVolumeOff
            : volume < 50
              ? IconVolumeDown
              : IconVolumeUp;

    return (
        <div className="flex h-8 items-center gap-1 rounded bg-black/30 px-2 transition-colors hover:bg-black/50 sm:h-10 sm:gap-2 sm:px-2">
            <div
                className="flex cursor-pointer items-center justify-center"
                onClick={() =>
                    handleChange(
                        volume === 0 ? lastNonZeroRef.current : 0,
                    )
                }
            >
                <VolumeIcon width="20px" height="20px" color="white" />
            </div>
            {/* Slider hidden on small screens — icon-only on mobile */}
            <div className="hidden w-16 items-center sm:flex sm:w-20">
                <Slider
                    value={[volume]}
                    onValueChange={(v) => handleChange(v[0])}
                    max={100}
                    step={1}
                    trackClassName="bg-gray-500/50"
                    rangeClassName="bg-white"
                    thumbClassName="border-white/50 bg-white"
                />
            </div>
        </div>
    );
}

function Combobox({
    options,
    value,
    onChange,
}: {
    options: string[];
    value: string;
    onChange: (value: string) => void;
}) {
    const [open, setOpen] = useState(false);
    const ref = useRef<HTMLDivElement>(null);

    const handleTriggerClick = (e: React.MouseEvent) => {
        if (ref.current && !ref.current.contains(e.target as Node)) {
            setOpen(false);
        } else {
            setOpen((prev) => !prev);
        }
    };

    return (
        <div
            ref={ref}
            className="relative flex cursor-pointer items-center justify-center"
        >
            <div
                className="flex h-8 cursor-pointer items-center justify-center rounded bg-black/30 px-2 text-xs text-white transition-colors hover:bg-black/50 sm:h-10 sm:px-3 sm:text-sm"
                onClick={handleTriggerClick}
            >
                {value}
            </div>
            {open && (
                <div className="absolute right-0 bottom-full mb-1 flex flex-col rounded-md bg-black/80 px-1 py-1">
                    {options.map((option) => (
                        <div
                            key={option}
                            className="flex cursor-pointer items-center gap-2 rounded border-0 px-1 text-white outline-none hover:bg-white/20"
                            onClick={() => {
                                onChange(option);
                                setOpen(false);
                            }}
                        >
                            <span className="w-4 shrink-0">
                                {value === option ? (
                                    <IconCheck color="white" />
                                ) : null}
                            </span>
                            <p className="w-14 text-xs sm:w-16 sm:text-sm">
                                {option}
                            </p>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
