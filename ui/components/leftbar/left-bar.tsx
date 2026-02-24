"use client";
import AllChannelsView from "./channels";
import { ResizablePanel } from "../ui/resizable";
import { useEffect, useState } from "react";
import IconToLeft from "../icons/to-left";
import { cn } from "@/utils/cn";
import IconToRight from "../icons/to-right";
import { Button } from "../ui/button";

export default function LeftBar() {
    const [minimizedLeftBar, setMinimizedLeftBar] = useState<boolean | null>(
        null,
    );

    useEffect(() => {
        const stored = localStorage.getItem("minimizeLeftBar");
        if (stored === "true") {
            setMinimizedLeftBar(true);
        } else {
            setMinimizedLeftBar(false);
        }
    }, []);

    if (minimizedLeftBar === null) return null;

    return (
        <ResizablePanel
            minSize={15}
            defaultSize={15}
            maxSize={25}
            id="1"
            order={1}
            className={cn(
                "bg-background relative h-full w-full min-w-[18rem] py-4",
                minimizedLeftBar ? "max-w-16 min-w-16" : "",
            )}
        >
            <AllChannelsView
                isMinimized={minimizedLeftBar}
                minimizeLeftBarIcon={MinimizeButton(
                    minimizedLeftBar,
                    setMinimizedLeftBar,
                )}
            />
        </ResizablePanel>
    );
}

const MinimizeButton = (
    minimizedLeftBar: boolean,
    setMinimizeLeftBar: (newVal: boolean) => any,
) => {
    const handleClick = (e: React.MouseEvent<HTMLElement>) => {
        e.preventDefault();
        e.stopPropagation();
        setMinimizeLeftBar(!minimizedLeftBar);
        localStorage.setItem(
            "minimizeLeftBar",
            JSON.stringify(!minimizedLeftBar),
        );
    };

    return (
        <Button
            className={cn(minimizedLeftBar ? "mx-auto" : "")}
            onClick={handleClick}
        >
            {minimizedLeftBar ? (
                <IconToRight className="stroke-primary-foreground" />
            ) : (
                <IconToLeft className="stroke-primary-foreground" />
            )}
        </Button>
    );
};
