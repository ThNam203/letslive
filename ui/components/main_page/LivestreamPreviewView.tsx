"use client"

import { ClassValue } from "clsx";
import { useRouter } from "next/navigation";
import { User } from "../../types/user";
import { Hover3DBox } from "../Hover3DBox";
import { cn } from "../../utils/cn";
import LivestreamPreviewDetailView from "./LivestreamPreviewDetailView";
import { LivestreamingPreview } from "../../types/livestreaming";
import GLOBAL from "../../global";

const LivestreamPreviewView = ({
    className,
    livestream,
}: {
    className?: ClassValue;
    livestream: LivestreamingPreview;
}) => {
    const router = useRouter();

    return (
        <div className={cn("flex flex-col gap-2 max-w-[300px]", className)}>
            <Hover3DBox
                viewers={0}
                showViewer={true}
                showStream={true}
                imageSrc={`${GLOBAL.API_URL}/files/livestreams/${livestream.id}/thumbnail.jpeg`}
                className="h-[170px] cursor-pointer"
                onClick={() => router.push(`/users/${livestream.userId}`)}
            />
            <LivestreamPreviewDetailView livestream={livestream}/>
        </div>
    );
};

export default LivestreamPreviewView;