import React from "react";
import { BaseIcon } from "@/components/icons/base-icon";
import { IconProp } from "@/types/icon-prop";

function IconChevronUp(props: IconProp) {
    return (
        <BaseIcon {...props}>
            <path
                fill="currentColor"
                d="M7.41 15.41L12 10.83l4.59 4.58L18 14l-6-6l-6 6z"
            />
        </BaseIcon>
    );
}

export default IconChevronUp;
