"use client";

import { useState } from "react";
import type { PanelImperativeHandle, PanelSize } from "react-resizable-panels";
import AllChannelsView from "./channels";
import { ResizablePanel } from "../ui/resizable";
import IconToLeft from "../icons/to-left";
import { cn } from "@/utils/cn";
import IconToRight from "../icons/to-right";
import { Button } from "../ui/button";
import { Dialog, DialogContent, DialogTitle } from "../ui/dialog";
import useT from "@/hooks/use-translation";
import {
    LEFT_BAR_COLLAPSED_WIDTH_PX,
    LEFT_BAR_MAX_WIDTH,
    LEFT_BAR_MIN_WIDTH_PX,
    LEFT_BAR_PANEL_ID,
    type LeftBarLayout,
} from "./left-bar-layout";

type LeftBarProps = {
    panelRef: React.RefObject<PanelImperativeHandle | null>;
    initialLayout: LeftBarLayout;
    isCollapsed: boolean;
    isSmallScreen: boolean;
    onResize: (size: PanelSize) => void;
    onToggleCollapse: () => void;
};

export default function LeftBar({
    panelRef,
    initialLayout,
    isCollapsed,
    isSmallScreen,
    onResize,
    onToggleCollapse,
}: LeftBarProps) {
    const [isOverlayOpen, setIsOverlayOpen] = useState(false);
    const { t } = useT("common");

    const isMinimized = isSmallScreen || isCollapsed;

    const handleToggle = (e: React.MouseEvent<HTMLElement>) => {
        e.preventDefault();
        e.stopPropagation();
        if (isSmallScreen) {
            setIsOverlayOpen(true);
            return;
        }
        onToggleCollapse();
    };

    return (
        <>
            <ResizablePanel
                id={LEFT_BAR_PANEL_ID}
                panelRef={panelRef}
                collapsible
                collapsedSize={LEFT_BAR_COLLAPSED_WIDTH_PX}
                minSize={LEFT_BAR_MIN_WIDTH_PX}
                maxSize={LEFT_BAR_MAX_WIDTH}
                defaultSize={
                    initialLayout.isCollapsed
                        ? LEFT_BAR_COLLAPSED_WIDTH_PX
                        : initialLayout.width
                }
                groupResizeBehavior="preserve-pixel-size"
                onResize={onResize}
                className="bg-background py-4"
            >
                <AllChannelsView
                    isMinimized={isMinimized}
                    minimizeLeftBarIcon={
                        <Button
                            className={cn(isMinimized ? "mx-auto" : "")}
                            onClick={handleToggle}
                        >
                            {isMinimized ? (
                                <IconToRight className="stroke-primary-foreground" />
                            ) : (
                                <IconToLeft className="stroke-primary-foreground" />
                            )}
                        </Button>
                    }
                />
            </ResizablePanel>

            <Dialog open={isOverlayOpen} onOpenChange={setIsOverlayOpen}>
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
                            <Button onClick={() => setIsOverlayOpen(false)}>
                                <IconToLeft className="stroke-primary-foreground" />
                            </Button>
                        }
                    />
                </DialogContent>
            </Dialog>
        </>
    );
}
