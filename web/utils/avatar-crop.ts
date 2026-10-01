import type { Area } from "react-easy-crop";
import {
    AVATAR_MAX_DIMENSION,
    AVATAR_MAX_ZOOM,
    AVATAR_MIN_DIMENSION,
} from "@/constant/image";
import type { AvatarCrop } from "@/types/user";

export type ImageSize = { width: number; height: number };

export function loadImage(src: string): Promise<HTMLImageElement> {
    return new Promise((resolve, reject) => {
        const img = new Image();
        img.onload = () => resolve(img);
        img.onerror = () => reject(new Error("failed to load image"));
        img.src = src;
    });
}

export function naturalSize(img: HTMLImageElement): ImageSize {
    return { width: img.naturalWidth, height: img.naturalHeight };
}

export function isAllowedAvatarSize({ width, height }: ImageSize): boolean {
    return (
        Math.min(width, height) >= AVATAR_MIN_DIMENSION &&
        Math.max(width, height) <= AVATAR_MAX_DIMENSION
    );
}

// at zoom 1 the crop square spans the image's short side; stop zooming
// before the square would cover fewer source pixels than the avatar has
export function avatarMaxZoom({ width, height }: ImageSize): number {
    const zoom = Math.min(width, height) / AVATAR_MIN_DIMENSION;
    return Math.min(AVATAR_MAX_ZOOM, Math.max(1, zoom));
}

// react-easy-crop rounds x, y, width and height separately, so the square can
// be off by a pixel or stick out of the image; the server rejects that
export function toAvatarCrop(area: Area, image: ImageSize): AvatarCrop {
    const size = Math.min(
        Math.round(Math.min(area.width, area.height)),
        image.width,
        image.height,
    );
    const clamp = (value: number, max: number) =>
        Math.min(Math.max(Math.round(value), 0), max);

    return {
        x: clamp(area.x, image.width - size),
        y: clamp(area.y, image.height - size),
        size,
    };
}

export function centerSquare({ width, height }: ImageSize): AvatarCrop {
    const size = Math.min(width, height);
    return {
        x: Math.floor((width - size) / 2),
        y: Math.floor((height - size) / 2),
        size,
    };
}

// returns a data URL of the crop scaled to outputSize x outputSize
export function renderCroppedSquare(
    img: HTMLImageElement,
    crop: AvatarCrop,
    outputSize: number,
    type = "image/png",
): string {
    const canvas = document.createElement("canvas");
    canvas.width = outputSize;
    canvas.height = outputSize;
    const ctx = canvas.getContext("2d");
    if (!ctx) throw new Error("canvas 2d context unavailable");

    ctx.imageSmoothingQuality = "high";
    ctx.drawImage(
        img,
        crop.x,
        crop.y,
        crop.size,
        crop.size,
        0,
        0,
        outputSize,
        outputSize,
    );
    return canvas.toDataURL(type);
}
