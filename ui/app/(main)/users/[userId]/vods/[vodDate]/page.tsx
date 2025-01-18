"use client";
import { StreamingFrame, VideoInfo } from "@/components/custom_react_player/streaming_frame";
import { VODFrame } from "@/components/custom_react_player/vod_frame";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";

export default function Livestreaming() {
  const params = useParams<{ userId: string, vodDate: string }>();
    const [playerInfo, setPlayerInfo] = useState<VideoInfo>({
      videoTitle: "Live Streaming",
      streamer: {
        name: "Dr. Pedophile",
      },
      videoUrl: null,
    })

  useEffect(() => {
    setPlayerInfo(prev => ({
        ...prev,
        videoUrl: `http://localhost:8889/static/${params.userId}/vods/${params.vodDate}/index.m3u8`
    }))
    console.log(`http://localhost:8889/static/${params.userId}/vods/${params.vodDate}/index.m3u8`);
  }, [params.userId, params.vodDate]);

  if (playerInfo.videoUrl === "") {
    return <div>Loading...</div>;
  }
  return <VODFrame videoInfo={playerInfo} />;
}