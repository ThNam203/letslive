import React from "react";
import { BaseIcon } from "@/components/icons/base-icon";
import { IconProp } from "@/types/icon-prop";

function IconRefresh(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path
                fill="currentColor"
                d="M17.65 6.35A7.96 7.96 0 0 0 12 4a8 8 0 0 0-8 8a8 8 0 0 0 8 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 12 18a6 6 0 0 1-6-6a6 6 0 0 1 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4z"
            />
        </BaseIcon>
    );
}

export default IconRefresh;
