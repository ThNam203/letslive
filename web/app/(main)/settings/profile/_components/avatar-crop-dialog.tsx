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
import {
    avatarMaxZoom,
    naturalSize,
    toCropSquare,
    type CropSquare,
} from "@/utils/avatar-crop";

type Props = {
    image: HTMLImageElement | null;
    onCancel: () => void;
    onApply: (square: CropSquare) => Promise<void>;
};

// Give each picked image its own key so it starts centered at zoom 1.
export default function AvatarCropDialog({ image, onCancel, onApply }: Props) {
    const { t } = useT(["settings", "common"]);
    const [crop, setCrop] = useState<Point>({ x: 0, y: 0 });
    const [zoom, setZoom] = useState(1);
    const [area, setArea] = useState<Area | null>(null);
    // react-easy-crop measures its size and position once, with
    // getBoundingClientRect, which includes the dialog's open animation
    // (scale and slide). Remounting when that animation ends makes it measure
    // the settled dialog; crop and zoom live here, so they survive.
    const [measureKey, setMeasureKey] = useState(0);
    const [isApplying, setIsApplying] = useState(false);

    const size = image ? naturalSize(image) : null;

    const handleApply = async () => {
        if (!size || !area) return;
        setIsApplying(true);
        try {
            await onApply(toCropSquare(area, size));
        } finally {
            setIsApplying(false);
        }
    };

    return (
        <Dialog
            open={image !== null}
            onOpenChange={(open) => {
                if (!open) onCancel();
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
                    <DialogTitle>
                        {t("settings:profile.crop_title")}
                    </DialogTitle>
                </DialogHeader>

                <div className="bg-muted relative h-72 w-full overflow-hidden rounded-md sm:h-80">
                    {image && size && (
                        <Cropper
                            key={measureKey}
                            image={image.src}
                            crop={crop}
                            zoom={zoom}
                            minZoom={1}
                            maxZoom={avatarMaxZoom(size)}
                            aspect={1}
                            cropShape="round"
                            showGrid={false}
                            onCropChange={setCrop}
                            onZoomChange={setZoom}
                            onCropComplete={(_, pixels) => setArea(pixels)}
                            mediaProps={{
                                alt: t("settings:profile.crop_image_alt"),
                            }}
                        />
                    )}
                </div>

                <DialogFooter>
                    <Button type="button" variant="outline" onClick={onCancel}>
                        {t("common:cancel")}
                    </Button>
                    <Button
                        type="button"
                        onClick={handleApply}
                        disabled={!area || isApplying}
                    >
                        {t("settings:profile.crop_apply")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
