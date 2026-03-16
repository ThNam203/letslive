"use client";

import AllChannelsView from "./channels";
import { ResizablePanel } from "../ui/resizable";
import { useEffect, useState } from "react";
import IconToLeft from "../icons/to-left";
import { cn } from "@/utils/cn";
import IconToRight from "../icons/to-right";
import { Button } from "../ui/button";

const MINIMIZED_LEFT_BAR_STATE_KEY = "isLeftBarMinimized";
const MINIMIZED_LEFT_BAR_STATE_DEFAULT = false;

export default function LeftBar() {
    const [minimizedLeftBar, setMinimizedLeftBar] = useState<boolean>(MINIMIZED_LEFT_BAR_STATE_DEFAULT);

    useEffect(() => {
        const stored = localStorage.getItem(MINIMIZED_LEFT_BAR_STATE_KEY);
        setMinimizedLeftBar(stored === "true");
    }, []);

    return (
        <ResizablePanel
            {...(minimizedLeftBar
                ? { minSize: 64, maxSize: 64, defaultSize: 64 }
                : {
                      minSize: "15%",
                      defaultSize: "20%",
                      maxSize: "35%",
                  })}
            id="1"
            className={"bg-background py-4"}
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
            MINIMIZED_LEFT_BAR_STATE_KEY,
            !minimizedLeftBar ? "true" : "false",
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
