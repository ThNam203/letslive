import React from "react";
import { BaseIcon } from "./base-icon";
import { IconProp } from "@/src/types/icon-prop";

function IconPause(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path fill="currentColor" d="M14 19h4V5h-4M6 19h4V5H6z" />
        </BaseIcon>
    );
}

export default IconPause;
