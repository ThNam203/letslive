import { ButtonProps } from "@/components/buttons/TagBtn";
import { cn } from "@/utils/cn";
import React, { ReactNode } from "react";

export interface IconButtonProps extends ButtonProps {
    icon: ReactNode;
}
const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>(
    ({ className, icon, ...props }, ref) => {
        return (
            <button
                ref={ref}
                className={cn(
                    "w-8 h-8 hover:bg-hoverColor disabled:hover:bg-transparent disabled:text-secondaryWord rounded flex flex-row items-center justify-center",
                    className
                )}
                {...props}
            >
                {icon}
            </button>
        );
    }
);
IconButton.displayName = "IconButton";

export default IconButton;
