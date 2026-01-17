import React from "react";
import { BaseIcon } from "./base-icon";
import { IconProp } from "@/src/types/icon-prop";

function IconClose(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path
                fill="currentColor"
                d="M19 6.41L17.59 5L12 10.59L6.41 5L5 6.41L10.59 12L5 17.59L6.41 19L12 13.41L17.59 19L19 17.59L13.41 12z"
            />
        </BaseIcon>
    );
}

export default IconClose;
