import { cn } from "@/utils/cn";
import { useRef } from "react";
import Description from "@/routes/(main)/settings/_components/description";
import ImageHover from "@/routes/(main)/settings/_components/image-hover";

type Props = {
    label: string;
    description?: string;
    onImageChange?: (file: File | null) => void;
    onResetImage?: () => void;
    imageUrl?: string | null;
    hoverText?: string;
    className?: string;
    showCloseIcon?: boolean;
};

export default function ImageField({
    label,
    description,
    className,
    onImageChange,
    onResetImage,
    imageUrl,
    hoverText,
    showCloseIcon = true,
}: Props) {
    const inputRef = useRef<HTMLInputElement>(null);
    const handleClick = () => {
        inputRef.current?.click(); // Trigger file input
    };

    return (
        <div className={className}>
            <label className="mb-2 block text-sm font-medium" htmlFor={label}>
                {label}
            </label>
            <div className="relative aspect-video w-full overflow-hidden rounded-lg">
                <div
                    className={cn(
                        "absolute h-full w-full cursor-pointer rounded-lg bg-cover bg-center bg-no-repeat",
                        !imageUrl && "border-border border border-dashed",
                    )}
                    style={{
                        backgroundImage: `${imageUrl ? `url("${imageUrl}")` : "none"}`,
                    }}
                />
                <ImageHover
                    id={label}
                    inputRef={inputRef}
                    onValueChange={onImageChange}
                    onClick={handleClick}
                    onCloseIconClick={onResetImage}
                    title={hoverText}
                    showCloseIcon={showCloseIcon}
                />
            </div>
            {description && (
                <Description content={description} className="mt-1" />
            )}
        </div>
    );
}
