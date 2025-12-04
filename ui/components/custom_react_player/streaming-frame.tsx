"use client";
import { Slider } from "@/components/ui/slider";
import { ClassValue } from "clsx";
import { useEffect, useRef, useState } from "react";
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

const playbackRates = {
    "0.5x": 0.5,
    "1x": 1.0,
    "1.5x": 1.5,
    "2x": 2.0,
};

export type VideoInfo = {
    videoUrl: string | null;
    videoTitle: string;
    streamer: {
        name: string;
    };
};

type Config = {
    playbackRate: number;
    resolution: string;
    volumeValue: number;
    isFullscreen: boolean;
    loop: boolean;
    pip: boolean;
};

type FnControl = {
    playVideo: () => void;
    pauseVideo: () => void;
    seekToTime: (time: number) => void;
    handleVolumeChange: (value: number) => void;
    onFullScreen: () => void;
    onExitFullScreen: () => void;
    handlePlaybackRateChange: (value: number) => void;
    handleResolutionChange: (value: string) => void;
};

export function StreamingFrame({
    videoInfo,
    className,
    onVideoStart,
}: {
    videoInfo: VideoInfo;
    onVideoStart?: () => void;
    className?: ClassValue;
}) {
    const playerRef = useRef<ReactPlayer>(null);

    const [count, setCount] = useState(0); // for hide control bar
    const [duration, setDuration] = useState(0);
    const [isPlaying, setIsPlaying] = useState(false);
    const [isLoading, setIsLoading] = useState(true); // for skeleton
    const [currentTime, setCurrentTime] = useState(0);
    const [loaded, setLoaded] = useState(0);
    const [config, setConfig] = useState<Config>({
        playbackRate: 1.0,
        resolution: "Auto",
        volumeValue: 100,
        isFullscreen: false,
        loop: false,
        pip: false, // Picture in Picture (the browser support this feature, not need to make it true)
    });
    const [resolutions, setResolutions] = useState<string[]>(["Auto"]);

    const [fnControl, setFnControl] = useState<FnControl>({
        playVideo: () => playVideo(),
        pauseVideo: () => pauseVideo(),
        seekToTime: (time: number) => seekToTime(time),
        handleVolumeChange: (value: number) => handleVolumeChange(value),
        onFullScreen: () => onFullScreen(),
        onExitFullScreen: () => onExitFullScreen(),
        handlePlaybackRateChange: (value: number) =>
            handlePlaybackRateChange(value),
        handleResolutionChange: (value: string) =>
            handleResolutionChange(value),
    });

    const handleVolumeChange = (value: number) => {
        setConfig({ ...config, volumeValue: value });
    };

    const playVideo = () => {
        setIsPlaying(true);
        if (onVideoStart && currentTime === 0) onVideoStart();
    };

    const pauseVideo = () => {
        setIsPlaying(false);
    };

    const seekToTime = (time: number) => {
        if (playerRef.current) {
            setCurrentTime(time);
            playerRef.current.seekTo(time);
        }
    };

    const onFullScreen = () => {
        const element = document.getElementById("frame");
        if (screenfull.isEnabled && element) {
            screenfull.request(element);
            setConfig({ ...config, isFullscreen: true });
        }
    };

    const onExitFullScreen = () => {
        if (screenfull.isEnabled) {
            screenfull.exit();
            setConfig({ ...config, isFullscreen: false });
        }
    };

    const handlePlaybackRateChange = (value: number) => {
        setConfig({ ...config, playbackRate: value });
    };

    const handleResolutionChange = (value: string) => {
        if (playerRef.current === null) return;
        if (!playerRef.current?.getInternalPlayer("hls")) return;

        setConfig((prevConfig) => ({ ...prevConfig, resolution: value }));

        if (value === "Auto") {
            playerRef.current.getInternalPlayer("hls").currentLevel = -1;
        } else {
            setResolutions((prevResolutions) => {
                // Extract height from the selected value (e.g., "1080p" -> 1080)
                const selectedHeight = getResolutionHeight(value);
                if (selectedHeight === null) {
                    return prevResolutions;
                }

                // Find the resolution that matches this height
                const levelIndex = prevResolutions.findIndex((reso) => {
                    if (reso === "Auto") return false;
                    const resoHeight = getResolutionHeight(reso);
                    return resoHeight !== null && resoHeight === selectedHeight;
                });

                if (levelIndex !== -1) {
                    playerRef.current!.getInternalPlayer("hls").currentLevel =
                        levelIndex - 1; // minus the "Auto" option
                }

                return prevResolutions;
            });
        }
    };

    return (
        <div
            className={cn("relative aspect-video w-full bg-black", className)}
            id="frame"
            onMouseMove={() => {
                setCount(0);
            }}
            onClick={() => {
                setCount(0);
            }}
        >
            {videoInfo.videoUrl != null && (
                <>
                    <ReactPlayerWrapper
                        playerRef={playerRef}
                        url={videoInfo.videoUrl}
                        muted={config.volumeValue === 0 ? true : false}
                        volume={config.volumeValue / 100}
                        playing={isPlaying}
                        width={"100%"}
                        height={"100%"}
                        playbackRate={config.playbackRate}
                        loop={config.loop}
                        onProgress={(state: any) => {
                            setCount(count + 1);
                            setLoaded(state.loaded);
                            setCurrentTime(state.playedSeconds);
                        }}
                        onDuration={(duration: number) => {
                            setDuration(duration);
                        }}
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
                    <FrontOfVideo
                        isPlaying={isPlaying}
                        currentTime={currentTime}
                        loaded={loaded}
                        config={config}
                        fnControl={fnControl}
                        duration={duration}
                        videoInfo={videoInfo}
                        resolutions={resolutions}
                        className={count > 3 ? "opacity-0" : "opacity-100"}
                    />
                </>
            )}
        </div>
    );
}

function FrontOfVideo({
    isPlaying,
    currentTime,
    loaded,
    config,
    fnControl,
    duration,
    videoInfo,
    resolutions,
    className,
}: {
    isPlaying: boolean;
    currentTime: number;
    loaded: number;
    config: Config;
    fnControl: FnControl;
    duration: number;
    videoInfo: VideoInfo;
    resolutions: string[];
    className?: ClassValue;
}) {
    return (
        <div
            className={cn(
                "absolute top-0 flex h-full w-full flex-col items-center justify-end border border-border",
                className,
            )}
        >
            <div className="absolute left-4 right-4 top-4 flex flex-row items-center justify-between font-sans font-bold text-white">
                <span className="rounded bg-black/70 px-2 py-1">
                    {videoInfo.videoTitle}
                </span>
                <span className="rounded bg-black/70 px-2 py-1">
                    {videoInfo.streamer.name}
                </span>
            </div>
            <div
                className="flex h-full w-full items-center justify-center"
                onMouseUp={() => {
                    if (isPlaying) {
                        if (fnControl.pauseVideo) fnControl.pauseVideo();
                    } else {
                        if (fnControl.playVideo) fnControl.playVideo();
                    }
                }}
            >
                {!isPlaying && (
                    <IconPlay
                        width="100px"
                        height="100px"
                        className="cursor-pointer"
                        color="white"
                    />
                )}
            </div>

            <VideoControl
                isPlaying={isPlaying}
                currentTime={currentTime}
                loaded={loaded}
                config={config}
                fnControl={fnControl}
                resolutions={resolutions}
                duration={duration}
            />
        </div>
    );
}

function VideoControl({
    isPlaying,
    currentTime,
    loaded,
    config,
    fnControl,
    duration,
    resolutions,
    className,
}: {
    isPlaying: boolean;
    currentTime: number;
    loaded: number;
    config: Config;
    fnControl: FnControl;
    duration: number;
    resolutions: string[];
    className?: ClassValue;
}) {
    return (
        <div
            className={cn(
                "flex h-fit w-full flex-col items-center justify-center px-10 pb-4 pt-4",
                className,
            )}
        >
            <VideoTracking
                className="w-full mb-3"
                isPlaying={isPlaying}
                currentTime={currentTime}
                loaded={loaded}
                config={config}
                fnControl={fnControl}
                duration={duration}
            />
            <VideoControlButtons
                className="w-full"
                isPlaying={isPlaying}
                currentTime={currentTime}
                duration={duration}
                config={config}
                fnControl={fnControl}
                resolutions={resolutions}
            />
        </div>
    );
}

function VideoTracking({
    className,
    isPlaying,
    currentTime,
    loaded,
    config,
    duration,
    fnControl,
}: {
    className?: ClassValue;
    isPlaying: boolean;
    currentTime: number;
    loaded: number;
    config: Config;
    duration: number;
    fnControl: FnControl;
}) {
    return (
        <div
            className={cn(
                "flex w-full items-center justify-center bg-transparent",
                className,
            )}
        >
            <Slider
                value={[duration !== 0 ? (currentTime / duration) * 100 : 0]}
                onValueChange={(value) => {
                    if (fnControl.seekToTime)
                        fnControl.seekToTime((value[0] / 100) * duration);
                }}
                max={100}
                step={0.1}
                trackClassName="bg-gray-500/50"
                rangeClassName="bg-white"
                thumbClassName="border-white/50 bg-white"
            />
        </div>
    );
}

function VideoControlButtons({
    className,
    isPlaying,
    currentTime,
    duration,
    config,
    fnControl,
    resolutions,
}: {
    className?: ClassValue;
    isPlaying: boolean;
    currentTime: number;
    duration: number;
    config: Config;
    fnControl: FnControl;
    resolutions: string[];
}) {
    const { t } = useT("common");

    return (
        <div
            className={cn(
                "flex w-full flex-row items-center justify-between text-white",
                className,
            )}
        >
            <div className="flex flex-row items-center gap-6">
                <div
                    onClick={() => {
                        if (isPlaying) {
                            if (fnControl.pauseVideo) fnControl.pauseVideo();
                        } else {
                            if (fnControl.playVideo) fnControl.playVideo();
                        }
                    }}
                    className="bg-black/30 rounded h-10 w-10 flex items-center justify-center hover:bg-black/50 transition-colors"
                >
                    {isPlaying ? (
                        <IconPause
                            width="24px"
                            height="24px"
                            className="cursor-pointer"
                            color="white"
                        />
                    ) : (
                        <IconPlay
                            width="24px"
                            height="24px"
                            className="cursor-pointer"
                            color="white"
                        />
                    )}
                </div>
                <VolumeButton onVolumeChange={fnControl.handleVolumeChange} />
                <div className="font-semibold text-red-600 flex flex-row items-center gap-2">
                    <div className="h-2 w-2 rounded-full bg-red-600"></div>
                    <span>{t("common:live")}</span>
                </div>
            </div>

            <div className="flex flex-row items-center gap-4">
                <Combobox
                    options={resolutions
                        .map((res) => formatResolutionForDisplay(res))
                        .filter((res): res is string => res !== null)}
                    value={formatResolutionForDisplay(config.resolution) || "Auto"}
                    onChange={fnControl.handleResolutionChange}
                />

                <div
                    onClick={() => {
                        if (screenfull.isFullscreen) {
                            if (fnControl.onExitFullScreen)
                                fnControl.onExitFullScreen();
                        } else {
                            if (fnControl.onFullScreen)
                                fnControl.onFullScreen();
                        }
                    }}
                    className="bg-black/30 rounded h-10 w-10 flex items-center justify-center hover:bg-black/50 transition-colors"
                >
                    {config.isFullscreen ? (
                        <IconFullscreenExit
                            width="24px"
                            height="24px"
                            className="cursor-pointer"
                            color="white"
                        />
                    ) : (
                        <IconFullscreen
                            width="24px"
                            height="24px"
                            className="cursor-pointer"
                            color="white"
                        />
                    )}
                </div>
            </div>
        </div>
    );
}

function VolumeButton({
    onVolumeChange,
}: {
    onVolumeChange: (value: number) => void;
}) {
    const [volumeValue, setVolumeValue] = useState(100);
    const [currentVolume, setCurrentVolume] = useState(100);

    const handleVolumeChange = (value: number) => {
        setVolumeValue(value);
        if (onVolumeChange) onVolumeChange(value);
    };

    useEffect(() => {
        if (volumeValue !== 0) setCurrentVolume(volumeValue);
    }, [volumeValue]);

    return (
        <div className="flex w-[120px] h-10 flex-row items-center gap-2 bg-black/30 rounded px-2 hover:bg-black/50 transition-colors">
            <div
                className="cursor-pointer text-white flex items-center justify-center"
                onClick={() => {
                    if (volumeValue === 0) handleVolumeChange(currentVolume);
                    else handleVolumeChange(0);
                }}
            >
                {volumeValue === 0 && <IconVolumeOff width="24px" height="24px" color="white" />}
                {volumeValue > 0 && volumeValue < 50 && (
                    <IconVolumeDown width="24px" height="24px" color="white" />
                )}
                {volumeValue >= 50 && <IconVolumeUp width="24px" height="24px" color="white" />}
            </div>
            <div className="flex-1 flex items-center">
                <Slider
                    value={[volumeValue]}
                    onValueChange={(value) => handleVolumeChange(value[0])}
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

const Combobox = ({
    options,
    value,
    onChange,
}: {
    options: string[];
    value: string;
    onChange: (value: string) => void;
}) => {
    const [showOptions, setShowOptions] = useState(false);
    const handleValueChange = (value: string) => {
        if (onChange) onChange(value);
        setShowOptions(false);
    };
    const ref = useRef<HTMLDivElement>(null);

    const handleClick = (e: any) => {
        if (ref.current && !ref.current.contains(e.target)) {
            setShowOptions(false);
        } else setShowOptions(!showOptions);
    };

    return (
        <div
            ref={ref}
            className="relative flex cursor-pointer flex-row items-center justify-center gap-4"
        >
            <div
                className={cn("cursor-pointer text-white bg-black/30 rounded h-10 flex items-center justify-center px-3 hover:bg-black/50 transition-colors")}
                onClick={(e: any) => handleClick(e)}
            >
                {value}
            </div>
            {showOptions && (
                <div className="absolute bottom-full flex h-fit w-fit flex-col items-center rounded-md bg-black/70 px-1 py-1">
                    {options.map((option) => (
                        <div
                            key={option}
                            className={cn(
                                "flex cursor-pointer flex-row items-center justify-start gap-2 rounded border-0 text-white outline-none hover:bg-white/20",
                            )}
                            onClick={() => handleValueChange(option)}
                        >
                            <span className="w-[20px]">
                                {value === option ? <IconCheck color="white" /> : null}
                            </span>
                            <p className="w-[70px]">{option}</p>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};
