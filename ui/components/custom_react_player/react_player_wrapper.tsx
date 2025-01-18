import { LegacyRef, RefObject } from "react";
import ReactPlayer, { ReactPlayerProps } from "react-player";

export default function ReactPlayerWrapper(
  props: ReactPlayerProps & {
    playerRef: RefObject<ReactPlayer>;
  }
) {
  return <ReactPlayer ref={props.playerRef} {...props} />;
}
