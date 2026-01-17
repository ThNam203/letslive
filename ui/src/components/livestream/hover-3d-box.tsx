"use client";

import { ClassValue } from "clsx";
import Link from "next/link";
import { cn } from "@/utils/cn";
import LiveImage from "@/components/livestream/live-image";

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
                "group relative z-0 aspect-video w-full bg-primary",
                className,
            )}
            onClick={onClick}
        >
            <div className="absolute left-0 top-0 h-full w-2 skew-y-[0deg] bg-primary duration-100 ease-linear group-hover:top-[-0.25rem] group-hover:skew-y-[-45deg]"></div>
            <div className="absolute bottom-0 right-0 h-2 w-full skew-x-[0deg] bg-primary duration-100 ease-linear group-hover:right-[-0.25rem] group-hover:skew-x-[-45deg]"></div>
            <LiveImage
                width={500}
                height={500}
                src={imageSrc}
                alt="Livestream preview image"
                className="absolute left-0 top-0 z-10 aspect-video w-full cursor-pointer duration-100 ease-linear group-hover:-translate-y-2 group-hover:translate-x-2"
                fallbackSrc={fallbackSrc || "/images/streaming.jpg"}
                refreshInterval={30000}
                alwaysRefresh={true}
            />
            <span
                className={cn(
                    "absolute left-2 top-2 z-20 rounded bg-red-600 p-1 text-white duration-100 ease-linear group-hover:-translate-y-2 group-hover:translate-x-2",
                    showStream ? "" : "hidden",
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
            className="text-primary underline-offset-2 opacity-100 hover:underline hover:opacity-90"
        >
            {content}
        </Link>
    );
};

export { CustomLink, Hover3DBox };
