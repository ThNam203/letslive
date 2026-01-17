import React from "react";
import { BaseIcon } from "./base-icon";
import { IconProp } from "@/types/icon-prop";

function IconVolumeDown(props: IconProp) {
    return (
        <BaseIcon {...props} viewBox="0 0 1200 1200">
            <path
                fill="currentColor"
                d="M170.671 384.805h243.816l366.784-298.94v1028.27l-366.784-298.94H170.671zm748.41-48.764q108.128 108.126 110.248 262.898q0 148.407-110.248 254.417l-74.205-76.325q76.326-76.325 76.325-180.212q0-106.008-76.325-184.453z"
            />
        </BaseIcon>
    );
}

export default IconVolumeDown;
