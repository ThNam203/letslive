import { cn } from "@/utils/cn";
import React from "react";

interface Props {
    content: string;
    className?: string;
}
export default function Description({ content, className }: Props) {
    return (
        <p className={cn("text-foreground-muted m-0 p-0 text-sm", className)}>
            {content}
        </p>
    );
}
