import React from "react";
import { BaseIcon } from "./base-icon";
import { IconProp } from "@/src/types/icon-prop";

function IconChevronDown(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path
                fill="currentColor"
                d="M7.41 8.58L12 13.17l4.59-4.59L18 10l-6 6l-6-6z"
            />
        </BaseIcon>
    );
}

export default IconChevronDown;
