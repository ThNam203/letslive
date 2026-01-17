import React from "react";
import { BaseIcon } from "./base-icon";
import { IconProp } from "@/src/types/icon-prop";

function IconFastForward(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path fill="currentColor" d="M13 6v12l8.5-6M4 18l8.5-6L4 6z" />
        </BaseIcon>
    );
}

export default IconFastForward;
