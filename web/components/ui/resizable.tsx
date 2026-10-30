"use client";

import {
    Group,
    Panel,
    Separator as SeparatorPrimitive,
} from "react-resizable-panels";

import { cn } from "@/utils/cn";

const ResizablePanelGroup = ({
    className,
    ...props
}: React.ComponentProps<typeof Group>) => (
    <Group className={cn("h-full w-full", className)} {...props} />
);

const ResizablePanel = Panel;

const ResizableHandle = ({
    className,
    ...props
}: React.ComponentProps<typeof SeparatorPrimitive>) => (
    <SeparatorPrimitive
        className={cn(
            "bg-border data-[separator=active]:bg-primary data-[separator=focus]:bg-primary data-[separator=hover]:bg-primary/60 relative w-px transition-colors outline-none after:absolute after:inset-y-0 after:left-1/2 after:w-1 after:-translate-x-1/2",
            className,
        )}
        {...props}
    />
);

export { ResizablePanelGroup, ResizablePanel, ResizableHandle };
