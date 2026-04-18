import * as React from "react";

import { cn } from "@/utils/cn";

import { Input, type InputProps } from "./input";

export type InputWithIconLabelProps = Omit<InputProps, "showCount"> & {
    icon: React.ReactNode;
    endAdornment?: React.ReactNode;
    rowClassName?: string;
    showCount?: boolean;
    /** When true, shows a destructive border and the character counter at max length even if `showCount` is false. */
    emitErrorSignalOnLimit?: boolean;
};

const InputWithIconLabel = React.forwardRef<
    HTMLInputElement,
    InputWithIconLabelProps
>(
    (
        {
            icon,
            showCount = false,
            emitErrorSignalOnLimit = false,
            maxLength,
            value,
            endAdornment,
            className,
            rowClassName,
            id,
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

        const iconSlot = (
            <label
                htmlFor={id ? id : undefined}
                className={cn(
                    "flex shrink-0",
                    id ? "cursor-pointer" : undefined,
                )}
            >
                {icon}
            </label>
        );

        return (
            <div className="w-full">
                <div
                    className={cn(
                        "border-border flex items-center gap-4 rounded-md border px-4",
                        limitError && "border-destructive",
                        rowClassName,
                    )}
                >
                    {iconSlot}
                    <Input
                        ref={ref}
                        id={id}
                        showCount={false}
                        maxLength={maxLength}
                        value={value}
                        className={cn(
                            "h-12 min-w-0 flex-1 border-none bg-transparent shadow-none focus-visible:ring-0",
                            className,
                        )}
                        {...props}
                    />
                    {endAdornment != null && (
                        <span className="flex shrink-0 items-center">
                            {endAdornment}
                        </span>
                    )}
                </div>
                {showCounter && (
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
                )}
            </div>
        );
    },
);
InputWithIconLabel.displayName = "InputWithIconLabel";

export { InputWithIconLabel };
