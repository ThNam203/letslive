import React from "react";
import { BaseIcon } from "./base-icon";
import { IconProp } from "@/types/icon-prop";

function IconShare(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <g
                fill="none"
                stroke="currentColor"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
            >
                <path d="m15 17l5-5l-5-5" />
                <path d="M4 18v-2a4 4 0 0 1 4-4h12" />
            </g>
        </BaseIcon>
    );
}

export default IconShare;
