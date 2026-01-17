import React from "react";
import { BaseIcon } from "./base-icon";
import { IconProp } from "@/types/icon-prop";

function IconToRight(props: IconProp) {
    return (
        <BaseIcon
            width="1.5rem"
            height="1.5rem"
            viewBox="0 0 48 48"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="4"
            {...props}
        >
            <g fill="none">
                <path d="M34 24.0083H6" />
                <path d="M22 12L34 24L22 36" />
                <path d="M42 12V36" />
            </g>
        </BaseIcon>
    );
}

export default IconToRight;
