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
        <div className={cn("flex flex-col gap-2 max-w-[300px]", className)}>
            <Hover3DBox
                viewers={viewers}
                showViewer={true}
                showStream={true}
                imageSrc={"/images/streaming.jpg"}
                className="h-[170px] cursor-pointer"
                onClick={() => router.push(`/users/${user.id}`)}
            />
            <LivestreamPreviewDetailView
                user={user}
                title={title}
                category={category}
            />
        </div>
    );
};

export default LivestreamPreviewView;