"use client";

import * as React from "react";
import * as TabsPrimitive from "@radix-ui/react-tabs";

import { cn } from "@/utils/cn";

const Tabs = TabsPrimitive.Root;

const TabsList = React.forwardRef<
    React.ElementRef<typeof TabsPrimitive.List>,
    React.ComponentPropsWithoutRef<typeof TabsPrimitive.List>
>(({ className, ...props }, ref) => (
    <TabsPrimitive.List
        ref={ref}
        className={cn(
            "text-muted-foreground bg-muted inline-flex h-9 items-center justify-center rounded-lg p-1",
            className,
        )}
        {...props}
    />
));
TabsList.displayName = TabsPrimitive.List.displayName;

const TabsTrigger = React.forwardRef<
    React.ElementRef<typeof TabsPrimitive.Trigger>,
    React.ComponentPropsWithoutRef<typeof TabsPrimitive.Trigger>
>(({ className, ...props }, ref) => (
    <TabsPrimitive.Trigger
        ref={ref}
        className={cn(
            "focus-visible:ring-ring ring-offset-background data-[state=active]:bg-background data-[state=active]:text-foreground inline-flex items-center justify-center rounded-md px-3 py-1 text-sm font-medium whitespace-nowrap transition-all focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50 data-[state=active]:shadow",
            className,
        )}
        {...props}
    />
));
TabsTrigger.displayName = TabsPrimitive.Trigger.displayName;

const TabsContent = React.forwardRef<
    React.ElementRef<typeof TabsPrimitive.Content>,
    React.ComponentPropsWithoutRef<typeof TabsPrimitive.Content>
>(({ className, ...props }, ref) => (
    <TabsPrimitive.Content
        ref={ref}
        className={cn(
            "focus-visible:ring-ring ring-offset-background mt-2 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none",
            className,
        )}
        {...props}
    />
));
TabsContent.displayName = TabsPrimitive.Content.displayName;

const UNDERLINE_TABS_LIST_CLASS =
    "border-border h-auto w-full justify-start gap-6 rounded-none border-b bg-transparent p-0";
const UNDERLINE_TABS_TRIGGER_CLASS =
    "text-muted-foreground hover:text-foreground data-[state=active]:text-foreground data-[state=active]:border-primary -mb-px rounded-none border-b-2 border-transparent bg-transparent px-1 pt-1 pb-3 text-sm font-semibold data-[state=active]:bg-transparent data-[state=active]:shadow-none";

export {
    Tabs,
    TabsList,
    TabsTrigger,
    TabsContent,
    UNDERLINE_TABS_LIST_CLASS,
    UNDERLINE_TABS_TRIGGER_CLASS,
};
