"use client";

import AllChannelsView from "./channels";
import { ResizablePanel } from "../ui/resizable";
import { useState } from "react";
import IconToLeft from "../icons/to-left";
import { cn } from "@/utils/cn";
import IconToRight from "../icons/to-right";
import { Button } from "../ui/button";
import {
    Dialog,
    DialogContent,
    DialogTitle,
} from "../ui/dialog";
import { MQ_MAX_MD } from "@/constant/breakpoints";
import useMediaQuery from "@/hooks/use-media-query";
import useT from "@/hooks/use-translation";

const MINIMIZED_LEFT_BAR_STATE_KEY = "isLeftBarMinimized";

function readMinimizedLeftBarFromStorage(): boolean {
    if (typeof window === "undefined") {
        return false;
    }
    return localStorage.getItem(MINIMIZED_LEFT_BAR_STATE_KEY) === "true";
}

export default function LeftBar() {
    const isSmallScreen = useMediaQuery(MQ_MAX_MD);
    const [minimizedLeftBar, setMinimizedLeftBar] = useState(
        readMinimizedLeftBarFromStorage,
    );
    const [overlayOpen, setOverlayOpen] = useState(false);
    const { t } = useT("common");

    const panelMinimized = isSmallScreen ? true : minimizedLeftBar;

    const handleToggle = (e: React.MouseEvent<HTMLElement>) => {
        e.preventDefault();
        e.stopPropagation();
        if (isSmallScreen) {
            setOverlayOpen(true);
            return;
        }
        const next = !minimizedLeftBar;
        setMinimizedLeftBar(next);
        localStorage.setItem(
            MINIMIZED_LEFT_BAR_STATE_KEY,
            next ? "true" : "false",
        );
    };

    return (
        <>
            <ResizablePanel
                {...(panelMinimized
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
                    isMinimized={panelMinimized}
                    minimizeLeftBarIcon={
                        <Button
                            className={cn(panelMinimized ? "mx-auto" : "")}
                            onClick={handleToggle}
                        >
                            {panelMinimized ? (
                                <IconToRight className="stroke-primary-foreground" />
                            ) : (
                                <IconToLeft className="stroke-primary-foreground" />
                            )}
                        </Button>
                    }
                />
            </ResizablePanel>

            <Dialog open={overlayOpen} onOpenChange={setOverlayOpen}>
                <DialogContent
                    showCloseButton={false}
                    className="data-[state=closed]:slide-out-to-left data-[state=open]:slide-in-from-left top-0 left-0 grid h-screen w-80 max-w-[85vw] translate-x-0 translate-y-0 gap-0 overflow-y-auto rounded-none border-r p-0 py-4 sm:rounded-none"
                >
                    <DialogTitle className="sr-only">
                        {t("common:channels")}
                    </DialogTitle>
                    <AllChannelsView
                        isMinimized={false}
                        minimizeLeftBarIcon={
                            <Button onClick={() => setOverlayOpen(false)}>
                                <IconToLeft className="stroke-primary-foreground" />
                            </Button>
                        }
                    />
                </DialogContent>
            </Dialog>
        </>
    );
}
