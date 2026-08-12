import React from "react";
import ReactPlayer from "react-player";

interface ReactPlayerWrapperProps
    extends React.ComponentProps<typeof ReactPlayer> {
    playerRef?: React.Ref<HTMLVideoElement>;
}

export default function ReactPlayerWrapper({
    playerRef,
    ...props
}: ReactPlayerWrapperProps) {
    return <ReactPlayer ref={playerRef} {...props} />;
}
