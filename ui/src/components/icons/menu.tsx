import { IconProp } from "@/types/icon-prop";
import { BaseIcon } from "./base-icon";

const IconMenu = (props: IconProp) => {
    return (
        <BaseIcon {...props}>
            <path
                fill="currentColor"
                d="M3 6h18v2H3zm0 5h18v2H3zm0 5h18v2H3z"
            />
        </BaseIcon>
    );
};

export default IconMenu;
