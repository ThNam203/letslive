import React from "react";
import { BaseIcon } from "./base-icon";
import { IconProp } from "@/types/icon-prop";

function IconPlay(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path fill="currentColor" d="M8 5.14v14l11-7z" />
        </BaseIcon>
    );
}

export default IconPlay;
