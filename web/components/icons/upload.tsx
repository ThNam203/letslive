import { IconProp } from "@/types/icon-prop";
import { BaseIcon } from "./base-icon";

const IconUpload = (props: IconProp) => {
    return (
        <BaseIcon {...props}>
            <path 
                fill="currentColor" 
                d="M9 10v6h6v-6h4l-7-7l-7 7zm3-4.2L14.2 8H13v6h-2V8H9.8zM19 18H5v2h14z"
            />
        </BaseIcon>
    );
};

export default IconUpload;
