import * as React from "react";

import { cn } from "@/utils/cn";

type InputProps = React.ComponentProps<"input"> & {
    showCount?: boolean;
    /** When true, shows a destructive border and the character counter at max length even if `showCount` is false. */
    emitErrorSignalOnLimit?: boolean;
};

const Input = React.forwardRef<HTMLInputElement, InputProps>(
    (
        {
            className,
            type,
            showCount = false,
            emitErrorSignalOnLimit = false,
            maxLength,
            value,
            ...props
        },
        ref,
    ) => {
        const currentLength = typeof value === "string" ? value.length : 0;
        const atLimit = maxLength !== undefined && currentLength >= maxLength;
        const limitError = emitErrorSignalOnLimit && atLimit;
        const showCounter =
            maxLength !== undefined &&
            (showCount || (emitErrorSignalOnLimit && atLimit));

        const inputElement = (
            <input
                type={type}
                maxLength={maxLength}
                value={value}
                className={cn(
                    "placeholder:text-muted-foreground border-border file:text-foreground flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-base shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
                    limitError && "border-destructive",
                    className,
                )}
                ref={ref}
                {...props}
            />
        );

        if (!showCounter) return inputElement;

        return (
            <div className="w-full">
                {inputElement}
                <span
                    className={cn(
                        "mt-1 block text-right text-xs",
                        limitError
                            ? "text-destructive"
                            : "text-muted-foreground",
                    )}
                >
                    {currentLength}/{maxLength}
                </span>
            </div>
        );
    },
);
Input.displayName = "Input";

export { Input };
export type { InputProps };
