"use client";

import { useEffect, useRef, useState } from "react";
import Image, { ImageProps } from "next/image";

interface LiveImageProps extends Omit<ImageProps, "src"> {
    src: string; // only allow string
    fallbackSrc: string;
    refreshInterval?: number; // in ms
    alwaysRefresh?: boolean;
}

export default function LiveImage({
    src,
    fallbackSrc,
    refreshInterval = 30_000,
    alwaysRefresh = true,
    ...props
}: LiveImageProps) {
    const [imgSrc, setImgSrc] = useState<string>(
        alwaysRefresh ? fallbackSrc : src,
    );
    const [lastFailed, setLastFailed] = useState<boolean>(false);
    const containerRef = useRef<HTMLDivElement | null>(null);
    const isVisibleRef = useRef<boolean>(true);

    useEffect(() => {
        if (!alwaysRefresh) return;

        const tryLoadImage = () => {
            const testImg = new window.Image();
            testImg.src = `${src}?t=${Date.now()}`;
            testImg.onload = () => {
                setImgSrc(testImg.src);
                setLastFailed(false);
            };
            testImg.onerror = () => {
                setImgSrc(fallbackSrc);
                setLastFailed(true);
            };
        };

        tryLoadImage();

        const interval = setInterval(() => {
            if (!isVisibleRef.current) return;
            if (alwaysRefresh || lastFailed) {
                tryLoadImage();
            }
        }, refreshInterval);

        return () => clearInterval(interval);
    }, [src, fallbackSrc, refreshInterval, alwaysRefresh, lastFailed]);

    useEffect(() => {
        if (!alwaysRefresh) return;
        const el = containerRef.current;
        if (!el || typeof IntersectionObserver === "undefined") return;
        const observer = new IntersectionObserver(
            (entries) => {
                isVisibleRef.current = entries.some((e) => e.isIntersecting);
            },
            { rootMargin: "100px" },
        );
        observer.observe(el);
        return () => observer.disconnect();
    }, [alwaysRefresh]);

    return (
        <div ref={containerRef} className="contents">
            <Image
                {...props}
                alt={props.alt ?? ""}
                src={imgSrc}
                unoptimized={alwaysRefresh}
                onError={
                    alwaysRefresh
                        ? undefined
                        : () => setImgSrc(fallbackSrc)
                }
            />
        </div>
    );
}
