import type { Area } from "react-easy-crop";
import { readFileAsDataUrl } from "@/utils/file";

export type ImageSize = { width: number; height: number };

// a rectangle in the image as the browser shows it (EXIF orientation applied)
export type CropRect = { x: number; y: number; width: number; height: number };

export type ExportImageSpec = {
    quality: number;
};

export type SourceImageSpec = ExportImageSpec & {
    sourceMaxDimension: number;
};

export type CropImageSpec = SourceImageSpec & {
    aspect: number;
    minCropWidth: number;
    maxZoom: number;
};

function loadImage(src: string): Promise<HTMLImageElement> {
    return new Promise((resolve, reject) => {
        const img = new Image();
        img.onload = () => resolve(img);
        img.onerror = () => reject(new Error("failed to load image"));
        img.src = src;
    });
}

export async function loadImageFile(file: Blob): Promise<HTMLImageElement> {
    return loadImage(await readFileAsDataUrl(file));
}

export function naturalSize(img: HTMLImageElement): ImageSize {
    return { width: img.naturalWidth, height: img.naturalHeight };
}

function fullRect({ width, height }: ImageSize): CropRect {
    return { x: 0, y: 0, width, height };
}

export function minCropSize({
    aspect,
    minCropWidth,
}: CropImageSpec): ImageSize {
    return {
        width: minCropWidth,
        height: Math.ceil(minCropWidth / aspect),
    };
}

export function fitsSourceLimit(
    { width, height }: ImageSize,
    { sourceMaxDimension }: SourceImageSpec,
): boolean {
    return Math.max(width, height) <= sourceMaxDimension;
}

export function fitsCrop(size: ImageSize, spec: CropImageSpec): boolean {
    const min = minCropSize(spec);
    return (
        size.width >= min.width &&
        size.height >= min.height &&
        fitsSourceLimit(size, spec)
    );
}

// at zoom 1 the crop spans the image's limiting side; zooming in past this
// would leave the crop narrower than minCropWidth source pixels
export function maxCropZoom(
    { width, height }: ImageSize,
    { aspect, minCropWidth, maxZoom }: CropImageSpec,
): number {
    const zoom = Math.min(width, height * aspect) / minCropWidth;
    return Math.min(maxZoom, Math.max(1, zoom));
}

// react-easy-crop rounds x, y, width and height separately, so the crop can
// be off by a pixel, off the aspect, or stick out of the image
export function toCropRect(
    area: Area,
    image: ImageSize,
    { aspect, minCropWidth }: CropImageSpec,
): CropRect {
    const width = Math.min(
        Math.max(Math.round(area.width), minCropWidth),
        image.width,
        Math.floor(image.height * aspect),
    );
    const height = Math.min(Math.round(width / aspect), image.height);
    const clamp = (value: number, max: number) =>
        Math.min(Math.max(Math.round(value), 0), max);

    return {
        x: clamp(area.x, image.width - width),
        y: clamp(area.y, image.height - height),
        width,
        height,
    };
}

function drawRect(
    img: HTMLImageElement,
    rect: CropRect,
    output: ImageSize,
    background?: string,
): HTMLCanvasElement {
    const canvas = document.createElement("canvas");
    canvas.width = output.width;
    canvas.height = output.height;
    const ctx = canvas.getContext("2d");
    if (!ctx) throw new Error("canvas 2d context unavailable");

    if (background) {
        ctx.fillStyle = background;
        ctx.fillRect(0, 0, output.width, output.height);
    }
    ctx.imageSmoothingQuality = "high";
    ctx.drawImage(
        img,
        rect.x,
        rect.y,
        rect.width,
        rect.height,
        0,
        0,
        output.width,
        output.height,
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

// Safari cannot encode WebP and hands back a PNG instead; JPEG has no alpha,
// hence the white background.
export async function exportImage(
    img: HTMLImageElement,
    rect: CropRect,
    { quality }: ExportImageSpec,
    fileName: string,
): Promise<File> {
    const output = { width: rect.width, height: rect.height };

    const webp = await canvasToBlob(
        drawRect(img, rect, output),
        "image/webp",
        quality,
    );
    if (webp.type === "image/webp") {
        return new File([webp], `${fileName}.webp`, { type: webp.type });
    }

    const jpeg = await canvasToBlob(
        drawRect(img, rect, output, "#ffffff"),
        "image/jpeg",
        quality,
    );
    return new File([jpeg], `${fileName}.jpg`, { type: jpeg.type });
}

export function exportWholeImage(
    img: HTMLImageElement,
    spec: ExportImageSpec,
    fileName: string,
): Promise<File> {
    return exportImage(img, fullRect(naturalSize(img)), spec, fileName);
}
