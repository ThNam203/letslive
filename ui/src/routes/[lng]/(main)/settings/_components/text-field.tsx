import { Input } from "@/components/ui/input";
import React, { ComponentProps } from "react";
import Description from "@/routes/[lng]/(main)/settings/_components/description";
import { cn } from "@/utils/cn";

type TextProps = ComponentProps<typeof Input>;
type Props = {
    label: string;
    description?: string;
} & TextProps;

export default function TextField({
    label,
    description,
    className,
    ...props
}: Props) {
    return (
        <div>
            <label className="mb-2 block text-sm font-medium" htmlFor={label}>
                {label}
            </label>
            <Input
                id={label}
                className={cn(
                    "border border-border text-foreground focus:outline-hidden",
                    className,
                )}
                {...props}
            />
            {description && (
                <Description content={description} className="mt-1" />
            )}
        </div>
    );
}
