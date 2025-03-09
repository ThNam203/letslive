
import { Heart } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { UserVOD } from "../types/user";

export default function VODLink({ item }: { item: UserVOD }) {
    return <div
    className="bg-gray-200 overflow-hidden shadow-sm rounded-sm"
>
    <Link
        className={`w-full h-[180px] inline-block hover:cursor-pointer`}
        href={`/users/${item.userId}/vods/${item.id}`}
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
            {item.title}
        </h3>
        <p className="text-sm text-gray-500 mt-1">
            {item.description && item.description.length > 50 ? `${item.description.substring(0, 47)}...` : item.description} • {datediffFromNow(item.endedAt)} ago
        </p>
        <div className="flex items-center mt-2 text-sm text-gray-500">
            <Heart className="w-4 h-4 mr-1" />
            <span>{item.viewCount} views</span>
        </div>
    </div>
</div>
}


function datediffFromNow(pastDate: string) {        
    const now = new Date();
    const past = new Date(pastDate);
    const seconds = Math.round((now.getTime() - past.getTime()) / 1000);

    if (seconds < 60) {
        return `${seconds} second${seconds !== 1 ? 's' : ''}`;
    }

    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) {
        return `${minutes} minute${minutes !== 1 ? 's' : ''}`;
    }

    const hours = Math.floor(minutes / 60);
    if (hours < 24) {
        return `${hours} hour${hours !== 1 ? 's' : ''}`;
    }

    const days = Math.floor(hours / 24);
    if (days < 7) {
        return `${days} day${days !== 1 ? 's' : ''}`;
    }

    const weeks = Math.floor(days / 7);
    if (days < 30) {
        return `${weeks} week${weeks !== 1 ? 's' : ''}`;
    }

    const months = Math.floor(days / 30);
    if (days < 365) {
        return `${months} month${months !== 1 ? 's' : ''}`;
    }

    const years = Math.floor(days / 365);
    return `${years} year${years !== 1 ? 's' : ''}`;
}