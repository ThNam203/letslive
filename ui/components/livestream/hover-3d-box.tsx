"use client";

import { ClassValue } from "clsx";
import Link from "next/link";
import { cn } from "../../utils/cn";
import LiveImage from "./live-image";

const Hover3DBox = ({
    className,
    imageSrc,
    fallbackSrc,
    showStream = false,
    onClick,
}: {
    className?: ClassValue;
    imageSrc: string;
    fallbackSrc?: string;
    showStream?: boolean;
    onClick?: () => void;
}) => {
    return (
        <div
            className={cn(
                "relative aspect-video w-full bg-primary z-0 group",
                className
            )}
            onClick={onClick}
        >
            <div className="absolute left-0 top-0 w-2 h-full skew-y-[0deg] bg-primary group-hover:skew-y-[-45deg] group-hover:top-[-0.25rem] ease-linear duration-100"></div>
            <div className="absolute bottom-0 right-0 w-full h-2 skew-x-[0deg] bg-primary group-hover:skew-x-[-45deg] group-hover:right-[-0.25rem] ease-linear duration-100"></div>
            <LiveImage
                width={500}
                height={500}
                src={imageSrc}
                alt="Livestream preview image"
                className="absolute aspect-video w-full top-0 left-0 z-10 group-hover:translate-x-2 group-hover:-translate-y-2 ease-linear duration-100 cursor-pointer"
                fallbackSrc={fallbackSrc || "/images/streaming.jpg"}
                refreshInterval={30000}
                alwaysRefresh={true}
            />
            <span
                className={cn(
                    "absolute text-white bg-red-600 rounded p-1 top-2 left-2 z-20 group-hover:translate-x-2 group-hover:-translate-y-2 ease-linear duration-100",
                    showStream ? "" : "hidden"
                )}
            >
                LIVE
            </span>
        </div>
    );
};

const CustomLink = ({ content, href }: { content: string; href: string }) => {
    return (
        <Link
            href={href}
            className="text-primary opacity-100 hover:opacity-90 hover:underline underline-offset-2"
        >
            {content}
        </Link>
    );
};

export { CustomLink, Hover3DBox };
