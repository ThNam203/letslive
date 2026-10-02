"use client";

import { useState } from "react";
import Cropper, { type Area, type Point } from "react-easy-crop";
import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import useT from "@/hooks/use-translation";
import { maxCropZoom, naturalSize, toCropRect } from "@/utils/image-crop";
import type { ImageCrop } from "./use-image-crop";

type Props = {
    crop: ImageCrop;
    title: string;
    cropShape: "rect" | "round";
};

// Key it with crop.dialogKey so each picked image starts centered at zoom 1.
export default function ImageCropDialog({ crop, title, cropShape }: Props) {
    const { image, spec } = crop;
    const { t } = useT(["settings", "common"]);
    const [position, setPosition] = useState<Point>({ x: 0, y: 0 });
    const [zoom, setZoom] = useState(1);
    const [area, setArea] = useState<Area | null>(null);
    // react-easy-crop measures its size and position once, with
    // getBoundingClientRect, which includes the dialog's open animation
    // (scale and slide). Remounting when that animation ends makes it measure
    // the settled dialog; position and zoom live here, so they survive.
    const [measureKey, setMeasureKey] = useState(0);
    const [isApplying, setIsApplying] = useState(false);

    const size = image ? naturalSize(image) : null;

    const handleApply = async () => {
        if (!size || !area) return;
        setIsApplying(true);
        try {
            await crop.apply(toCropRect(area, size, spec));
        } finally {
            setIsApplying(false);
        }
    };

    return (
        <Dialog
            open={image !== null}
            onOpenChange={(open) => {
                if (!open) crop.cancel();
            }}
        >
            <DialogContent
                className="sm:max-w-md"
                aria-describedby={undefined}
                onAnimationEnd={(event) => {
                    if (event.target === event.currentTarget)
                        setMeasureKey((key) => key + 1);
                }}
            >
                <DialogHeader>
                    <DialogTitle>{title}</DialogTitle>
                </DialogHeader>

                <div className="bg-muted relative h-72 w-full overflow-hidden rounded-md sm:h-80">
                    {image && size && (
                        <Cropper
                            key={measureKey}
                            image={image.src}
                            crop={position}
                            zoom={zoom}
                            minZoom={1}
                            maxZoom={maxCropZoom(size, spec)}
                            aspect={spec.aspect}
                            cropShape={cropShape}
                            showGrid={false}
                            onCropChange={setPosition}
                            onZoomChange={setZoom}
                            onCropComplete={(_, pixels) => setArea(pixels)}
                            mediaProps={{
                                alt: t("settings:crop.image_alt"),
                            }}
                        />
                    )}
                </div>

                <DialogFooter>
                    <Button
                        type="button"
                        variant="outline"
                        onClick={crop.cancel}
                    >
                        {t("common:cancel")}
                    </Button>
                    <Button
                        type="button"
                        onClick={handleApply}
                        disabled={!area || isApplying}
                    >
                        {t("settings:crop.apply")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
