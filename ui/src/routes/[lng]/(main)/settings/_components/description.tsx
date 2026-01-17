import { cn } from "@/src/utils/cn";
import React from "react";

interface Props {
    content: string;
    className?: string;
}
export default function Description({ content, className }: Props) {
    return (
        <p className={cn("m-0 p-0 text-sm text-foreground-muted", className)}>
            {content}
        </p>
    );
}
