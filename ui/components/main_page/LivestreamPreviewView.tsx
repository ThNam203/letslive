"use client";

import { ClassValue } from "clsx";
import { useRouter } from "next/navigation";
import { User } from "../../types/user";
import { Hover3DBox } from "../Hover3DBox";
import { cn } from "../../utils/cn";
import LivestreamPreviewDetailView from "./LivestreamPreviewDetailView";
import GLOBAL from "../../global";
import { Livestream } from "../../types/livestream";
import { useEffect, useState } from "react";
import { GetUserById } from "../../lib/api/user";
import { Card, CardContent } from "../ui/card";

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
        <Card className="w-full transition-all hover:shadow-md rounded-sm">
            <Hover3DBox
                viewers={0}
                showViewer={true}
                showStream={true}
                imageSrc={
                    livestream.thumbnailUrl ??
                    `${GLOBAL.API_URL}/files/livestreams/${livestream.id}/thumbnail.jpeg`
                }
                className="h-[207px] cursor-pointer mb-4"
                onClick={() => router.push(`/users/${livestream.userId}`)}
            />
            <CardContent>
                <LivestreamPreviewDetailView
                    livestream={livestream}
                    user={user}
                />
            </CardContent>
        </Card>
    );
};

export default LivestreamPreviewView;
