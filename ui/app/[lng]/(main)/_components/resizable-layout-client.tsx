"use client";

import {
    ResizableHandle,
    ResizablePanel,
    ResizablePanelGroup,
} from "@/components/ui/resizable";
import LeftBar from "@/components/leftbar/left-bar";
import { useDefaultLayout } from "react-resizable-panels";

export function ResizableLayoutClient({
    children,
}: {
    children: React.ReactNode;
}) {
    const { defaultLayout, onLayoutChanged } = useDefaultLayout({
        id: "sidebar-size",
        storage: typeof window !== "undefined" ? localStorage : undefined,
    });
    return (
        <ResizablePanelGroup
            defaultLayout={defaultLayout}
            onLayoutChanged={onLayoutChanged}
            orientation="horizontal"
            className="flex-1"
        >
            <LeftBar />
            <ResizableHandle />
            <ResizablePanel
                id="2"
                className="bg-background min-h-0 overflow-hidden"
            >
                {children}
            </ResizablePanel>
        </ResizablePanelGroup>
    );
}
