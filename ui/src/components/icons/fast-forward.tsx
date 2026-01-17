import React from "react";
import { BaseIcon } from "@/components/icons/base-icon";
import { IconProp } from "@/types/icon-prop";

function IconFastForward(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path fill="currentColor" d="M13 6v12l8.5-6M4 18l8.5-6L4 6z" />
        </BaseIcon>
    );
}

export default IconFastForward;
