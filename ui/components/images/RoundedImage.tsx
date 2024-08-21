import { cn } from "@/utils/cn";
import Image, { ImageProps } from "next/image";
import React from "react";

export interface RoundedImageProps extends ImageProps {}

const RoundedImage = React.forwardRef<HTMLImageElement, RoundedImageProps>(
    ({ className, ...props }, ref) => {
        return (
            <Image
                ref={ref}
                className={cn(
                    "w-8 h-8 rounded-full flex flex-row items-center justify-center outline-none",
                    className
                )}
                {...props}
            ></Image>
        );
    }
);
RoundedImage.displayName = "RoundedImage";

export default RoundedImage;
