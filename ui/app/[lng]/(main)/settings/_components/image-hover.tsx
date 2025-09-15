import IconClose from "@/components/icons/close";
import { FILE_SIZE_LIMIT_MB_UNIT } from "@/constant/image";
import { cn } from "@/utils/cn";
import { IsValidFileSizeInMB } from "@/utils/file";
import React from "react";
import { toast } from "react-toastify";

type InputProps = React.InputHTMLAttributes<HTMLInputElement>;

type Props = {
  className?: string;
  onClick?: () => void;
  inputRef?: React.RefObject<HTMLInputElement | null>;
  onValueChange?: (file: File) => void;
  title?: string | React.ReactNode;
  showCloseIcon?: boolean;
  closeIconPosition?: "top" | "bottom" | "top-right";
  onCloseIconClick?: () => void;
} & InputProps;

export default function ImageHover({
  className,
  onClick,
  inputRef,
  onValueChange,
  title,
  showCloseIcon = true,
  closeIconPosition = "top-right",
  onCloseIconClick,
  ...props
}: Props) {
  const handleValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      if (!IsValidFileSizeInMB(file, FILE_SIZE_LIMIT_MB_UNIT)) {
        toast.error(`File size exceeds ${FILE_SIZE_LIMIT_MB_UNIT} MB`);
        return;
      }
      onValueChange?.(file);
    }
  };

  const postions = {
    top: "absolute left-1/2 -translate-x-1/2 top-2",
    bottom: "absolute left-1/2 -translate-x-1/2 bottom-2",
    "top-right": "absolute right-2 top-2",
  };

  const handleCloseIconClick = (event: React.MouseEvent) => {
    event.stopPropagation();
    if (inputRef?.current) inputRef.current.value = "";
    onCloseIconClick?.();
  };

  return (
    <div
      className={cn(
        "absolute w-full h-full flex items-center justify-center bg-background-hover/20 opacity-0 hover:opacity-100 transition-all duration-300 ease-in-out cursor-pointer",
        className
      )}
      onClick={onClick}
    >
      <input
        type="file"
        ref={inputRef}
        className="hidden"
        onChange={handleValueChange}
        {...props}
      />
      <div className="flex flex-row gap-2 text-foreground">{title}</div>
      <div
        className={cn(
          "p-1 flex items-center justify-center rounded-full bg-primary cursor-pointer",
          postions[closeIconPosition],
          !showCloseIcon && "hidden"
        )}
        onClick={handleCloseIconClick}
      >
        <IconClose className="text-primary-foreground" />
      </div>
    </div>
  );
}
