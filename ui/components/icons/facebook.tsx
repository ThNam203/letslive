import { BaseIcon } from "./base-icon";
import { IconProp } from "@/types/icon-prop";

function IconFacebook(props: IconProp) {
    return (
        <BaseIcon viewBox="0 0 24 24" {...props}>
            <path 
                fill="currentColor" 
                d="M9.198 21.5h4v-8.01h3.604l.396-3.98h-4V7.5a1 1 0 0 1 1-1h3v-4h-3a5 5 0 0 0-5 5v2.01h-2l-.396 3.98h2.396z" 
            />
        </BaseIcon>
    );
}

export default IconFacebook;
