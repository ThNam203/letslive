import type { Area } from "react-easy-crop";

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

export function isWithinDimensions(
    { width, height }: ImageSize,
    minDimension: number,
    maxDimension: number,
): boolean {
    return (
        Math.min(width, height) >= minDimension &&
        Math.max(width, height) <= maxDimension
    );
}

// at zoom 1 the crop square spans the image's short side; zooming in past
// this would leave the square covering fewer than minCropSize source pixels
export function maxCropZoom(
    { width, height }: ImageSize,
    minCropSize: number,
    zoomCap: number,
): number {
    const zoom = Math.min(width, height) / minCropSize;
    return Math.min(zoomCap, Math.max(1, zoom));
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

function canvasToBlob(
    canvas: HTMLCanvasElement,
    type: string,
    quality: number,
): Promise<Blob> {
    return new Promise((resolve, reject) => {
        canvas.toBlob(
            (blob) =>
                blob
                    ? resolve(blob)
                    : reject(new Error("canvas export failed")),
            type,
            quality,
        );
    });
}

type ExportSquareOptions = {
    maxDimension: number;
    quality: number;
    fileName: string;
};

// Safari cannot encode WebP and hands back a PNG instead; JPEG has no alpha,
// hence the white background.
export async function exportSquare(
    img: HTMLImageElement,
    square: CropSquare,
    { maxDimension, quality, fileName }: ExportSquareOptions,
): Promise<File> {
    const size = Math.min(square.size, maxDimension);

    const webp = await canvasToBlob(
        drawSquare(img, square, size),
        "image/webp",
        quality,
    );
    if (webp.type === "image/webp") {
        return new File([webp], `${fileName}.webp`, { type: webp.type });
    }

    const jpeg = await canvasToBlob(
        drawSquare(img, square, size, "#ffffff"),
        "image/jpeg",
        quality,
    );
    return new File([jpeg], `${fileName}.jpg`, { type: jpeg.type });
}
