import React from "react";
import { BaseIcon } from "@/components/icons/base-icon";
import { IconProp } from "@/types/icon-prop";

function IconCopy(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path
                fill="currentColor"
                d="M19 21H8V7h11m0-2H8a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h11a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2m-3-4H4a2 2 0 0 0-2 2v14h2V3h12z"
            />
        </BaseIcon>
    );
}

export default IconCopy;
