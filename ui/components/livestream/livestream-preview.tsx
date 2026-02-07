"use client";

import { ClassValue } from "clsx";
import { useRouter } from "next/navigation";
import { PublicUser } from "../../types/user";
import { Hover3DBox } from "./hover-3d-box";
import LivestreamPreviewDetailView from "./livestream-preview-detail";
import GLOBAL from "../../global";
import { Livestream } from "../../types/livestream";
import { useEffect, useState } from "react";
import { GetUserById } from "../../lib/api/user";
import { Card, CardContent } from "../ui/card";
import { cn } from "@/utils/cn";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";

const LivestreamPreviewView = ({
    className,
    livestream,
}: {
    className?: ClassValue;
    livestream: Livestream;
}) => {
    const router = useRouter();
    const [user, setUser] = useState<PublicUser | null>(null);
    const { t } = useT(["api-response", "fetch-error"]);

    useEffect(() => {
        const fetchUserInfo = async () => {
            await GetUserById(livestream.userId)
                .then((res) => {
                    if (res.success) {
                        setUser(res.data ?? null);
                    } else {
                        toast(t(`api-response:${res.key}`), {
                            toastId: res.requestId,
                            type: "error",
                        });
                    }
                })
                .catch((_) => {
                    toast(t("fetch-error:client_fetch_error"), {
                        toastId: "client-fetch-error-id",
                        type: "error",
                    });
                });
        };

        fetchUserInfo();
    }, [livestream]);

    return (
        <Card
            className={cn(
                "w-full rounded-sm border-muted transition-all hover:shadow-md",
                className,
            )}
        >
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
            <CardContent className="bg-muted p-4">
                <LivestreamPreviewDetailView
                    livestream={livestream}
                    user={user}
                />
            </CardContent>
        </Card>
    );
};

export default LivestreamPreviewView;
