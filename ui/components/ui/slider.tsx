"use client";

import * as React from "react";
import * as SliderPrimitive from "@radix-ui/react-slider";
import { cn } from "@/utils/cn";

interface SliderProps
    extends React.ComponentPropsWithoutRef<typeof SliderPrimitive.Root> {
    trackClassName?: string;
    rangeClassName?: string;
    thumbClassName?: string;
}

const Slider = React.forwardRef<
    React.ElementRef<typeof SliderPrimitive.Root>,
    SliderProps
>(
    (
        { className, trackClassName, rangeClassName, thumbClassName, ...props },
        ref,
    ) => (
        <SliderPrimitive.Root
            ref={ref}
            className={cn(
                "relative flex w-full touch-none items-center select-none",
                className,
            )}
            {...props}
        >
            <SliderPrimitive.Track
                className={cn(
                    "bg-primary/20 relative h-1.5 w-full grow overflow-hidden rounded-full",
                    trackClassName,
                )}
            >
                <SliderPrimitive.Range
                    className={cn("bg-primary absolute h-full", rangeClassName)}
                />
            </SliderPrimitive.Track>
            <SliderPrimitive.Thumb
                className={cn(
                    "focus-visible:ring-ring border-primary/50 bg-background block h-4 w-4 rounded-full border shadow transition-colors focus-visible:ring-1 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50",
                    thumbClassName,
                )}
            />
        </SliderPrimitive.Root>
    ),
);
Slider.displayName = SliderPrimitive.Root.displayName;

export { Slider };
