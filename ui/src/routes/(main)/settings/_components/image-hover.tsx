import IconClose from "@/components/icons/close";
import { FILE_SIZE_LIMIT_MB_UNIT } from "@/constant/image";
import { cn } from "@/utils/cn";
import { IsValidFileSizeInMB } from "@/utils/file";
import React from "react";
import { toast } from "react-toastify";
import useT from "@/hooks/use-translation";

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
    const { t } = useT("settings");
    const handleValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (file) {
            if (!IsValidFileSizeInMB(file, FILE_SIZE_LIMIT_MB_UNIT)) {
                toast.error(
                    t("settings:file_size_exceeds", {
                        size: FILE_SIZE_LIMIT_MB_UNIT,
                    }),
                );
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
                "bg-background-hover/20 absolute flex h-full w-full cursor-pointer items-center justify-center opacity-0 transition-all duration-300 ease-in-out hover:opacity-100",
                className,
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
            <div className="text-foreground flex flex-row gap-2">{title}</div>
            <div
                className={cn(
                    "bg-primary flex cursor-pointer items-center justify-center rounded-full p-1",
                    postions[closeIconPosition],
                    !showCloseIcon && "hidden",
                )}
                onClick={handleCloseIconClick}
            >
                <IconClose className="text-primary-foreground" />
            </div>
        </div>
    );
}
