
import { Heart } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { datediffFromNow, formatSeconds } from "../utils/timeFormats";
import { Livestream } from "../types/livestream";
import GLOBAL from "../global";

export default function VODLink({ vod, classname }: { vod: Livestream, classname?: string }) {
    return <div
    className={`bg-gray-200 overflow-hidden shadow-sm rounded-sm ${classname}`}
>
    <Link
        className={`w-full inline-block hover:cursor-pointer`}
        href={`/users/${vod.userId}/vods/${vod.id}`}
    >
        <div className="flex flex-col items-center justify-center h-full bg-black bg-opacity-50 aspect-video">
            <Image
                alt="vod icon"
                src={vod.thumbnailUrl ?? `${GLOBAL.API_URL}/files/livestreams/${vod.id}/thumbnail.jpeg`}
                width={0}
                height={0}
                sizes="100vw"
                className="aspect-video w-full"
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
            <span>{vod.viewCount} {vod.viewCount < 2 ? "view" : "views"}</span>
        </div>
    </div>
</div>
}
