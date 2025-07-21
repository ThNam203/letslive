import React from "react";
import { BaseIcon } from "./base-icon";
import { IconProp } from "@/types/icon-prop";

function IconSend(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path fill="currentColor" d="M3 20v-6l8-2l-8-2V4l19 8Z" />
        </BaseIcon>
    );
}

export default IconSend;
