import * as React from "react";

import { cn } from "@/utils/cn";

type TextareaProps = React.ComponentProps<"textarea"> & {
    showCount?: boolean;
};

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
    ({ className, showCount, maxLength, value, ...props }, ref) => {
        const textareaElement = (
            <textarea
                maxLength={maxLength}
                value={value}
                className={cn(
                    "border-input placeholder:text-muted-foreground flex min-h-[60px] w-full rounded-md border bg-transparent px-3 py-2 text-base shadow-sm focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
                    className,
                )}
                ref={ref}
                {...props}
            />
        );

        if (!showCount || maxLength === undefined) return textareaElement;

        const currentLength = typeof value === "string" ? value.length : 0;

        return (
            <div className="w-full">
                {textareaElement}
                <span className="text-muted-foreground mt-1 block text-right text-xs">
                    {currentLength}/{maxLength}
                </span>
            </div>
        );
    },
);
Textarea.displayName = "Textarea";

export { Textarea };
export type { TextareaProps };
