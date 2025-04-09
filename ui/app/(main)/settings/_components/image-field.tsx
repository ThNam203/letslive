import { cn } from "@/utils/cn";
import { useRef } from "react";
import Description from "./description";
import ImageHover from "./image-hover";

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
      <label className="block text-sm font-medium mb-2" htmlFor={label}>
        {label}
      </label>
      <div className="relative w-full aspect-video rounded-lg overflow-hidden">
        <div
          className={cn(
            "absolute w-full h-full rounded-lg cursor-pointer bg-cover bg-center bg-no-repeat",
            !imageUrl && "border-2 border-dashed border-gray-300"
          )}
          style={{
            backgroundImage: `url("${imageUrl}")`,
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
      {description && <Description content={description} className="mt-1" />}
    </div>
  );
}
