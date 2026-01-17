import React from "react";
import ReactPlayer from "react-player";

interface ReactPlayerWrapperProps
    extends React.ComponentProps<typeof ReactPlayer> {
    playerRef?: React.Ref<ReactPlayer>;
}

export default function ReactPlayerWrapper({
    playerRef,
    ...props
}: ReactPlayerWrapperProps) {
    return <ReactPlayer ref={playerRef} {...props} />;
}
