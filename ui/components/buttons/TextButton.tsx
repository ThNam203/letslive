import { cn } from "@/utils/cn";
import React, { ReactNode } from "react";

export interface TextButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    iconBefore?: ReactNode;
    content?: string;
    iconAfter?: ReactNode;
  }

  const TextButton = React.forwardRef<HTMLButtonElement, TextButtonProps>(
    ({ className, content, iconBefore, iconAfter, children, ...props }, ref) => {
      return (
        <button
          ref={ref}
          className={cn(
            "px-2 py-2 bg-gray-200 hover:bg-hoverColor disabled:bg-primary/60 rounded text-xs font-bold text-gray-500 flex flex-row items-center justify-center gap-2 ease-linear duration-100 cursor-pointer disabled:cursor-default",
            className
          )}
          {...props}
        >
          {iconBefore}
          {content}
          {children}
          {iconAfter}
        </button>
      );
    }
  );
  TextButton.displayName = "TextButton";

    export default TextButton;