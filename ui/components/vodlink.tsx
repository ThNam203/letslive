
import { Heart } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { datediffFromNow, formatSeconds } from "../utils/timeFormats";
import { Livestream } from "../types/livestream";

export default function VODLink({ vod, classname }: { vod: Livestream, classname?: string }) {
    return <div
    className={`bg-gray-200 overflow-hidden shadow-sm rounded-sm ${classname}`}
>
    <Link
        className={`w-full h-[180px] inline-block hover:cursor-pointer`}
        href={`/users/${vod.userId}/vods/${vod.id}`}
    >
        <div className="flex flex-col items-center justify-center h-full bg-black bg-opacity-50">
            <Image
                alt="vod icon"
                src={"/icons/video.svg"}
                width={100}
                height={100}
            />
            {/* <p className="text-white">
                Streamed on {item}
            </p> */}
        </div>
    </Link>
    <div className="p-4">
        <h3 className="font-semibold text-gray-900">
            {vod.title}
        </h3>
        <p className="text-sm text-gray-500 mt-1">
            {formatSeconds(vod.duration)}
        </p>
        <p className="text-sm text-gray-500 mt-1">
            {vod.description && vod.description.length > 50 ? `${vod.description.substring(0, 47)}...` : vod.description} • {datediffFromNow(vod.endedAt)} ago
        </p>
        <div className="flex items-center mt-2 text-sm text-gray-500">
            <Heart className="w-4 h-4 mr-1" />
            <span>{vod.viewCount} views</span>
        </div>
    </div>
</div>
}
