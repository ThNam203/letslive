"use client";

import { useState } from "react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { toast } from "@/components/utils/toast";
import { formatTime } from "@/components/custom_react_player/video-player";
import IconCopy from "@/components/icons/copy";
import useT from "@/hooks/use-translation";

export default function ShareVODDialog({
    open,
    onOpenChange,
    vodPath,
    currentTime,
}: {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    /** e.g. /users/{userId}/vods/{vodId} */
    vodPath: string;
    /** whole seconds into the video when the dialog was opened */
    currentTime: number;
}) {
    const { t } = useT("common");
    const [startAtCurrent, setStartAtCurrent] = useState(false);

    const canStartAt = currentTime > 0;
    const url = new URL(vodPath, window.location.origin);
    if (canStartAt && startAtCurrent) {
        url.searchParams.set("t", String(currentTime));
    }
    const link = url.toString();

    const copyLink = async () => {
        try {
            await navigator.clipboard.writeText(link);
            toast.success(t("common:vod.link_copied"));
        } catch {
            toast.error(t("common:vod.copy_failed"));
        }
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-md">
                <DialogHeader>
                    <DialogTitle>{t("common:vod.share")}</DialogTitle>
                    <DialogDescription>
                        {t("common:vod.share_description")}
                    </DialogDescription>
                </DialogHeader>

                <div className="flex items-center gap-2">
                    <Input
                        value={link}
                        readOnly
                        aria-label={t("common:vod.share_link")}
                        onFocus={(e) => e.currentTarget.select()}
                        className="flex-1"
                    />
                    <Button onClick={copyLink} className="shrink-0">
                        <IconCopy />
                        {t("common:vod.copy")}
                    </Button>
                </div>

                {canStartAt && (
                    <div className="flex items-center gap-2">
                        <Switch
                            id="share-start-at"
                            checked={startAtCurrent}
                            onCheckedChange={setStartAtCurrent}
                        />
                        <Label htmlFor="share-start-at">
                            {t("common:vod.start_at", {
                                time: formatTime(currentTime),
                            })}
                        </Label>
                    </div>
                )}
            </DialogContent>
        </Dialog>
    );
}
