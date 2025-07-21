import React from "react";
import { BaseIcon } from "./base-icon";
import { IconProp } from "@/types/icon-prop";

function IconPlay(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path
                fill="currentColor"
                d="M18.4 12.5L9 18.38L8 19V6zm-1.9 0L9 7.8v9.4z"
            />
        </BaseIcon>
    );
}

export default IconPlay;
