import Link from "next/link";
import { dateDiffFromNow, formatSeconds } from "../../utils/timeFormats";
import GLOBAL from "../../global";
import { VOD } from "@/types/vod";
import LiveImage from "./live-image";
import { cn } from "@/utils/cn";

export default function VODView({
    vod,
    classname,
}: {
    vod: VOD;
    classname?: string;
}) {
    return (
        <div
            className={cn(
                `overflow-hidden rounded-sm bg-background shadow-sm`,
                classname,
            )}
        >
            <Link
                className={`block w-full hover:cursor-pointer`}
                href={`/users/${vod.userId}/vods/${vod.id}`}
            >
                <div className="flex aspect-video h-full flex-col items-center justify-center bg-black bg-opacity-50">
                    <LiveImage
                        alt="vod icon"
                        src={
                            vod.thumbnailUrl ??
                            `${GLOBAL.API_URL}/files/livestreams/${vod.id}/thumbnail.jpeg`
                        }
                        width={0}
                        height={0}
                        sizes="100vw"
                        className="aspect-video w-full"
                        fallbackSrc="/images/streaming.jpg"
                        alwaysRefresh={false}
                    />
                </div>
            </Link>

            <div className="border border-border p-4 border-t-0">
                <h3 className="font-semibold text-foreground">{vod.title}</h3>
                <p className="mt-1 text-sm text-foreground-muted">
                    {formatSeconds(vod.duration)}
                </p>
                <p className="mt-1 text-sm text-foreground-muted">
                    {vod.description && vod.description.length > 50
                        ? `${vod.description.substring(0, 47)}...`
                        : vod.description}{" "}
                    • {dateDiffFromNow(vod.createdAt)} ago
                </p>
                <div className="mt-2 flex items-center text-sm text-foreground-muted">
                    <span>
                        {vod.viewCount} {vod.viewCount < 2 ? "view" : "views"}
                    </span>
                </div>
            </div>
        </div>
    );
}
