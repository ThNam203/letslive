"use client"

import { ClassValue } from "clsx";
import { useRouter } from "next/navigation";
import { User } from "../../types/user";
import { Hover3DBox } from "../Hover3DBox";
import { cn } from "../../utils/cn";
import LivestreamPreviewDetailView from "./LivestreamPreviewDetailView";

const LivestreamPreviewView = ({
    className,
    viewers,
    title,
    category,
    tags,
    user,
}: {
    className?: ClassValue;
    viewers: number;
    title: string;
    tags: string[];
    category?: string;
    user: User;
}) => {
    const router = useRouter();

    return (
        <div className={cn("flex flex-col gap-2", className)}>
            {/* <Hover3DBox
                viewers={viewers}
                showViewer={true}
                showStream={true}
                imageSrc={}
                className="h-[170px]"
                onClick={() => router.push(`/users/${user.id}`)}
            /> */}
            <LivestreamPreviewDetailView
                user={user}
                title={title}
                category={category}
                tags={tags}
            />
        </div>
    );
};

export default LivestreamPreviewView;