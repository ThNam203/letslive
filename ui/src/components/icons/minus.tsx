import React from "react";
import { BaseIcon } from "./base-icon";
import { IconProp } from "@/types/icon-prop";

function IconMinus(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path fill="currentColor" d="M19 13H5v-2h14z" />
        </BaseIcon>
    );
}

export default IconMinus;
