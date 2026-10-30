"use client";

import { useRef, useState } from "react";
import {
    usePanelRef,
    type LayoutChangedMeta,
    type PanelSize,
} from "react-resizable-panels";
import {
    ResizableHandle,
    ResizablePanel,
    ResizablePanelGroup,
} from "@/components/ui/resizable";
import LeftBar from "@/components/leftbar/left-bar";
import {
    LEFT_BAR_COLLAPSED_WIDTH_PX,
    saveLeftBarLayout,
    type LeftBarLayout,
} from "@/components/leftbar/left-bar-layout";
import { MQ_MAX_MD } from "@/constant/breakpoints";
import useMediaQuery from "@/hooks/use-media-query";

export function MainBodyPanels({
    initialLeftBarLayout,
    children,
}: {
    initialLeftBarLayout: LeftBarLayout;
    children: React.ReactNode;
}) {
    const leftBarRef = usePanelRef();
    const isSmallScreen = useMediaQuery(MQ_MAX_MD);
    const [isLeftBarCollapsed, setIsLeftBarCollapsed] = useState(
        initialLeftBarLayout.isCollapsed,
    );
    const savedLayoutRef = useRef(initialLeftBarLayout);

    const persistLeftBarLayout = (layout: LeftBarLayout) => {
        savedLayoutRef.current = layout;
        saveLeftBarLayout(layout);
    };

    const handleLeftBarResize = (size: PanelSize) => {
        setIsLeftBarCollapsed(
            Math.round(size.inPixels) <= LEFT_BAR_COLLAPSED_WIDTH_PX,
        );
    };

    const handleLayoutChanged = (_: unknown, meta: LayoutChangedMeta) => {
        const panel = leftBarRef.current;
        if (!panel || !meta.isUserInteraction) return;
        const isCollapsed = panel.isCollapsed();
        persistLeftBarLayout({
            width: isCollapsed
                ? savedLayoutRef.current.width
                : panel.getSize().inPixels,
            isCollapsed,
        });
    };

    const toggleLeftBar = () => {
        const panel = leftBarRef.current;
        if (!panel) return;
        const { width } = savedLayoutRef.current;
        if (panel.isCollapsed()) {
            panel.resize(width);
            persistLeftBarLayout({ width, isCollapsed: false });
        } else {
            panel.collapse();
            persistLeftBarLayout({ width, isCollapsed: true });
        }
    };

    return (
        <ResizablePanelGroup
            id="main-body"
            onLayoutChanged={handleLayoutChanged}
        >
            <LeftBar
                panelRef={leftBarRef}
                initialLayout={initialLeftBarLayout}
                isCollapsed={isLeftBarCollapsed}
                isSmallScreen={isSmallScreen}
                onResize={handleLeftBarResize}
                onToggleCollapse={toggleLeftBar}
            />
            <ResizableHandle disabled={isSmallScreen} />
            <ResizablePanel id="content" className="bg-background min-h-0">
                {children}
            </ResizablePanel>
        </ResizablePanelGroup>
    );
}
