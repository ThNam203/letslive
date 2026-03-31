import * as React from "react";

import { cn } from "@/utils/cn";

type InputProps = React.ComponentProps<"input"> & {
    showCount?: boolean;
};

const Input = React.forwardRef<HTMLInputElement, InputProps>(
    ({ className, type, showCount, maxLength, value, ...props }, ref) => {
        const inputElement = (
            <input
                type={type}
                maxLength={maxLength}
                value={value}
                className={cn(
                    "placeholder:text-muted-foreground border-border file:text-foreground flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-base shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
                    className,
                )}
                ref={ref}
                {...props}
            />
        );

        if (!showCount || maxLength === undefined) return inputElement;

        const currentLength = typeof value === "string" ? value.length : 0;

        return (
            <div className="w-full">
                {inputElement}
                <span className="text-muted-foreground mt-1 block text-right text-xs">
                    {currentLength}/{maxLength}
                </span>
            </div>
        );
    },
);
Input.displayName = "Input";

export { Input };
export type { InputProps };
