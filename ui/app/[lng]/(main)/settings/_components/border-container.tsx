import { cn } from "@/utils/cn";
import React, { ReactNode } from "react";

interface Props {
    children: ReactNode | ReactNode[];
    className?: string;
}
export default function BorderContainer({ children, className }: Props) {
    return (
        <div
            className={cn(
                "border-border relative w-full rounded-lg border p-6",
                className,
            )}
        >
            {children}
        </div>
    );
}
