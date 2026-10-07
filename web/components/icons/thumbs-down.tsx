import React from "react";
import { BaseIcon } from "./base-icon";
import { IconProp } from "@/types/icon-prop";

function IconThumbsDown({
    filled = false,
    ...props
}: IconProp & { filled?: boolean }) {
    return (
        <BaseIcon {...props}>
            <g
                fill={filled ? "currentColor" : "none"}
                stroke="currentColor"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
            >
                <path d="M17 14V2" />
                <path d="M9 18.12L10 14H4.17a2 2 0 0 1-1.92-2.56l2.33-8A2 2 0 0 1 6.5 2H20a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2h-2.76a2 2 0 0 0-1.79 1.11L12 22a3.13 3.13 0 0 1-3-3.88Z" />
            </g>
        </BaseIcon>
    );
}

export default IconThumbsDown;
