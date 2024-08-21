"use client";

import { cn } from "@/utils/cn";
import React from "react";

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {}

export interface TagButtonProps extends ButtonProps {
  content: string;
}

const TagButton = React.forwardRef<HTMLButtonElement, TagButtonProps>(
  ({ className, content, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(
          "px-2 py-1 bg-gray-200 hover:bg-hoverColor disabled:hover:bg-transparent rounded-xl text-xs font-semibold text-gray-500 flex flex-row items-center justify-center cursor-pointer",
          className
        )}
        {...props}
      >
        {content}
      </button>
    );
  }
);
TagButton.displayName = "TagButton";

export default TagButton;