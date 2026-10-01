import type { Area } from "react-easy-crop";
import {
    AVATAR_MAX_DIMENSION,
    AVATAR_MAX_ZOOM,
    AVATAR_MIN_DIMENSION,
    AVATAR_UPLOAD_MAX_SIZE,
    AVATAR_UPLOAD_QUALITY,
} from "@/constant/image";

export type ImageSize = { width: number; height: number };

// a square in the image as the browser shows it (EXIF orientation applied)
export type CropSquare = { x: number; y: number; size: number };

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
// be off by a pixel or stick out of the image
export function toCropSquare(area: Area, image: ImageSize): CropSquare {
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

export function centerSquare({ width, height }: ImageSize): CropSquare {
    const size = Math.min(width, height);
    return {
        x: Math.floor((width - size) / 2),
        y: Math.floor((height - size) / 2),
        size,
    };
}

export function drawSquare(
    img: HTMLImageElement,
    square: CropSquare,
    outputSize: number,
    background?: string,
): HTMLCanvasElement {
    const canvas = document.createElement("canvas");
    canvas.width = outputSize;
    canvas.height = outputSize;
    const ctx = canvas.getContext("2d");
    if (!ctx) throw new Error("canvas 2d context unavailable");

    if (background) {
        ctx.fillStyle = background;
        ctx.fillRect(0, 0, outputSize, outputSize);
    }
    ctx.imageSmoothingQuality = "high";
    ctx.drawImage(
        img,
        square.x,
        square.y,
        square.size,
        square.size,
        0,
        0,
        outputSize,
        outputSize,
    );
    return canvas;
}

function canvasToBlob(canvas: HTMLCanvasElement, type: string): Promise<Blob> {
    return new Promise((resolve, reject) => {
        canvas.toBlob(
            (blob) =>
                blob
                    ? resolve(blob)
                    : reject(new Error("canvas export failed")),
            type,
            AVATAR_UPLOAD_QUALITY,
        );
    });
}

// Safari cannot encode WebP and hands back a PNG instead; JPEG has no alpha,
// hence the white background.
export async function exportAvatar(
    img: HTMLImageElement,
    square: CropSquare,
): Promise<File> {
    const size = Math.min(square.size, AVATAR_UPLOAD_MAX_SIZE);

    const webp = await canvasToBlob(
        drawSquare(img, square, size),
        "image/webp",
    );
    if (webp.type === "image/webp") {
        return new File([webp], "avatar.webp", { type: webp.type });
    }

    const jpeg = await canvasToBlob(
        drawSquare(img, square, size, "#ffffff"),
        "image/jpeg",
    );
    return new File([jpeg], "avatar.jpg", { type: jpeg.type });
}
