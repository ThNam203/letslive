"use client";
import {
    Check,
    Fullscreen,
    FullscreenExit,
    Pause,
    PlayArrow,
    VolumeDown,
    VolumeOff,
    VolumeUp,
} from "@mui/icons-material";
import { Slider } from "@mui/material";
import { ClassValue } from "clsx";
import { useEffect, useRef, useState } from "react";
import ReactPlayer from "react-player";
import screenfull from "screenfull";
import { cn } from "@/utils/cn";

import dynamic from "next/dynamic";
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

const RESOLUTION_TO_CLASS: { [key: string]: number } = {
    "416x234": 240,
    "640x360": 360,
    "768x432": 480,
    "960x540": 576,
    "1280x720": 720,
    "1920x1080": 1080,
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
        // more info
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
                const levelIndex = prevResolutions.findIndex((reso) => {
                    return (
                        parseInt(value.replace("p", "")) ===
                        RESOLUTION_TO_CLASS[reso]
                    );
                });

                playerRef.current!.getInternalPlayer("hls").currentLevel =
                    levelIndex - 1; // minus the "Auto" option

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
                    <PlayArrow
                        sx={{ fontSize: 100 }}
                        className="cursor-pointer text-white"
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
                "flex h-fit w-full flex-col items-center justify-center bg-black/60 px-10 pb-4 pt-4",
                className,
            )}
        >
            <VideoTracking
                className="w-full"
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
                value={duration !== 0 ? (currentTime / duration) * 100 : 0}
                onChange={(e: any) => {
                    if (fnControl.seekToTime)
                        fnControl.seekToTime((e.target.value / 100) * duration);
                }}
                size="small"
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
                >
                    {isPlaying ? (
                        <Pause
                            sx={{ fontSize: 24 }}
                            className="cursor-pointer text-white"
                        />
                    ) : (
                        <PlayArrow
                            sx={{ fontSize: 24 }}
                            className="cursor-pointer text-white"
                        />
                    )}
                </div>
                <VolumeButton onVolumeChange={fnControl.handleVolumeChange} />
                {/* <span className="text-white">
          {formatTime(currentTime)} / {formatTime(duration)}
        </span> */}
                <div className="font-semibold text-red-600">
                    <div className="h-2 w-2 rounded-full bg-red-600"></div>
                    Live
                </div>
            </div>

            <div className="flex flex-row items-center gap-4">
                {/* <Combobox
          options={Object.keys(playbackRates)}
          value={config.playbackRate + "x"}
          onChange={(value: string) =>
            fnControl.handlePlaybackRateChange(
              playbackRates[value as keyof typeof playbackRates]
            )
          }
        /> */}

                <Combobox
                    options={resolutions.map((res) =>
                        res === "Auto"
                            ? "Auto"
                            : RESOLUTION_TO_CLASS[res] + "p",
                    )}
                    value={config.resolution}
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
                >
                    {config.isFullscreen ? (
                        <FullscreenExit
                            sx={{ fontSize: 24 }}
                            className="cursor-pointer text-white"
                        />
                    ) : (
                        <Fullscreen
                            sx={{ fontSize: 24 }}
                            className="cursor-pointer text-white"
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
        <div className="flex w-[120px] flex-row items-center gap-4">
            <div
                className="cursor-pointer text-white"
                onClick={() => {
                    if (volumeValue === 0) handleVolumeChange(currentVolume);
                    else handleVolumeChange(0);
                }}
            >
                {volumeValue === 0 && <VolumeOff sx={{ fontSize: 24 }} />}
                {volumeValue > 0 && volumeValue < 50 && (
                    <VolumeDown sx={{ fontSize: 24 }} />
                )}
                {volumeValue >= 50 && <VolumeUp sx={{ fontSize: 24 }} />}
            </div>
            <Slider
                value={volumeValue}
                onChange={(e: any) => handleVolumeChange(e.target.value)}
                size="small"
            />
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
                className={cn("cursor-pointer text-white")}
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
                                {value === option ? <Check /> : null}
                            </span>
                            <p className="w-[70px]">{option}</p>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};
