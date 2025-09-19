import { Ref, RefObject } from "react";
import ReactPlayer, { ReactPlayerProps } from "react-player";

export default function ReactPlayerWrapper(
  props: ReactPlayerProps & {
    playerRef: RefObject<ReactPlayer | null>;
  }
) {
  return <ReactPlayer ref={props.playerRef} {...props} />;
}
