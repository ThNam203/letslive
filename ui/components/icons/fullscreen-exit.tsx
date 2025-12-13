import React from "react";
import { BaseIcon } from "./base-icon";
import { IconProp } from "@/types/icon-prop";

function IconFullscreenExit(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path
                fill="currentColor"
                d="M14 14h5v2h-3v3h-2zm-9 0h5v5H8v-3H5zm3-9h2v5H5V8h3zm11 3v2h-5V5h2v3z"
            />
        </BaseIcon>
    );
}

export default IconFullscreenExit;
