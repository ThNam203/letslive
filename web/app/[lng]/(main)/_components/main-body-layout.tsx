"use client";

import {
    ResizableHandle,
    ResizablePanel,
    ResizablePanelGroup,
} from "@/components/ui/resizable";
import LeftBar from "@/components/leftbar/left-bar";

export function MainBodyLayout({ children }: { children: React.ReactNode }) {
    return (
        <ResizablePanelGroup orientation="horizontal">
            <LeftBar />
            <ResizableHandle />
            <ResizablePanel id="2" className="bg-background min-h-0">
                {children}
            </ResizablePanel>
        </ResizablePanelGroup>
    );
}
