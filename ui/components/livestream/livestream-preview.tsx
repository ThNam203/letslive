"use client";

import { ClassValue } from "clsx";
import { useParams, useRouter } from "next/navigation";
import { User } from "../../types/user";
import { Hover3DBox } from "./hover-3d-box";
import LivestreamPreviewDetailView from "./livestream-preview-detail";
import GLOBAL from "../../global";
import { Livestream } from "../../types/livestream";
import { useEffect, useState } from "react";
import { GetUserById } from "../../lib/api/user";
import { Card, CardContent } from "../ui/card";
import { cn } from "@/utils/cn";

const LivestreamPreviewView = ({
    className,
    livestream,
}: {
    className?: ClassValue;
    livestream: Livestream;
}) => {
    const router = useRouter();
    const [user, setUser] = useState<User | null>(null);

    useEffect(() => {
        const fetchUserInfo = async () => {
            const { user } = await GetUserById(livestream.userId);
            if (user) {
                setUser(user);
            }
        };

        fetchUserInfo();
    }, [livestream]);

    return (
        <Card className={cn("w-full transition-all hover:shadow-md rounded-sm border-muted", className)}>
            <Hover3DBox
                showStream={true}
                imageSrc={
                    livestream.thumbnailUrl ??
                    `${GLOBAL.API_URL}/files/livestreams/${livestream.id}/thumbnail.jpeg`
                }
                fallbackSrc="/images/streaming.jpg"
                className="cursor-pointer"
                onClick={() => router.push(`/users/${livestream.userId}`)}
            />
            <CardContent className="p-4 bg-muted">
                <LivestreamPreviewDetailView
                    livestream={livestream}
                    user={user}
                />
            </CardContent>
        </Card>
    );
};

export default LivestreamPreviewView;
