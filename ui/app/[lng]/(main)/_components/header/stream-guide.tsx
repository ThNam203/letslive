"use client";
import { cn } from "@/utils/cn";
import React, { useEffect } from "react";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "../../../../../components/ui/popover";
import IconLiveStream from "../../../../../components/icons/live-stream";
import { Button } from "../../../../../components/ui/button";
import useT from "@/hooks/use-translation";

export default function StreamGuide() {
    const [open, setOpen] = React.useState(false);
    const [isGotIt, setIsGotIt] = React.useState(true);
    const { t } = useT("common");

    useEffect(() => {
        const readStreamGuiding = localStorage.getItem("readStreamGuiding");
        setIsGotIt(readStreamGuiding === "true");
    }, []);

    const handleGotIt = () => {
        setIsGotIt(true);
        setOpen(false);
        localStorage.setItem("readStreamGuiding", JSON.stringify(true)); // keep it true forever
    };

    return (
        <Popover open={open} onOpenChange={setOpen}>
            <PopoverTrigger>
                <div
                    className={cn(
                        "bg-whitecursor-pointer animate-bounce transition-all duration-700",
                        open && "scale-100 animate-none",
                        isGotIt && "animate-none",
                    )}
                >
                    <IconLiveStream className="text-primary" />
                </div>
            </PopoverTrigger>
            <PopoverContent className="border-border bg-muted text-foreground mr-4 w-100 border text-sm">
                <h1 className="text-xl font-semibold">
                    {t("how_to_livestream")}
                </h1>
                <p>{t("stream_guide_step_1")}</p>
                <p>
                    {t("stream_guide_step_2", {
                        server_url: `rtmp://${
                            process.env.NEXT_PUBLIC_ENVIRONMENT === "production"
                                ? "sen1or-huly.com"
                                : "localhost"
                        }:1935`,
                        interpolation: { escapeValue: false },
                    })}
                </p>
                <p className="mb-2">{t("start_your_livestream")}</p>
                <div className="flex w-full flex-row justify-center">
                    <Button onClick={handleGotIt}>{t("got_it")}</Button>
                </div>
            </PopoverContent>
        </Popover>
    );
}
