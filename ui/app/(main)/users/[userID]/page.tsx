"use client";
import IconButton from "@/components/buttons/IconBtn";
import TextButton from "@/components/buttons/TextButton";
import {
  StreamingFrame,
  VideoInfo,
} from "@/components/custom_react_player/streaming_frame";
import { InputWithIcon } from "@/components/Input";
import { User } from "@/models/User";
import { format } from "date-fns";
import { Smile } from "lucide-react";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";

export default function Livestreaming() {
  const [videoStream, setVideoStream] = useState<MediaStream | null>(null);
  const [micStream, setMicStream] = useState<MediaStream | null>(null);
  const [mixedStream, setMixedStream] = useState<MediaStream | null>(null);
  const [recorder, setRecorder] = useState<MediaRecorder | null>(null);
  const params = useParams<{ userID: string }>();
  const [playerInfo, setPlayerInfo] = useState<VideoInfo>({
    videoTitle: "Live Streaming",
    streamer: {
      name: "Dr. Pedophile",
    },
    videoUrl: null,
  })

  const [timeVideoStart, setTimeVideoStart] = useState<Date>(new Date());
  const [chatMessage, setChatMessage] = useState("");

  useEffect(() => {
    setPlayerInfo(prev => ({
        ...prev,
        videoUrl: `http://localhost:8889/static/${params.userID}/index.m3u8`
    }))
  }, [params.userID]);

  const setupStreams = async () => {
    const videoStream = await navigator.mediaDevices.getDisplayMedia({
      video: true,
      audio: {
        noiseSuppression: true,
        echoCancellation: true,
        sampleRate: 44100,
      },
    });

    const micStream = await navigator.mediaDevices.getUserMedia({
      audio: {
        noiseSuppression: true,
        echoCancellation: true,
        sampleRate: 44100,
      },
    });

    videoStream.getVideoTracks()[0].onended = () => {
      alert("Your screen sharing has ended");
      setVideoStream(null);
      setMicStream(null);
      setMixedStream(null);

      if (recorder) {
        recorder.stop();
        setRecorder(null);
      }
    };

    setVideoStream(videoStream);
    setMicStream(micStream);
  };

  const getTimeBaseOnVideo = (timeUserChat: Date, timeVideoStart: Date) => {
    const timeChat = timeUserChat.getTime() / 1000;
    const timeStart = timeVideoStart.getTime() / 1000;
    return Math.floor(timeChat - timeStart);
  };

  return (
    <div className="w-full h-full overflow-hidden flex lg:flex-row max-lg:flex-col">
      <div className="w-full h-[36vw] max-lg:shrink-0">
        <StreamingFrame
          videoInfo={playerInfo}
          onVideoStart={() => {
            setTimeVideoStart(new Date());
          }}
        />
      </div>
      <div className="lg:w-[400px] max-lg:w-full h-full font-sans border-l flex flex-col justify-between">
      </div>
    </div>
  );
}
