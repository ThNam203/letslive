import { useRef, useState } from "react";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import { readFileAsDataUrl } from "@/utils/file";
import {
    exportImage,
    fitsCrop,
    loadImageFile,
    minCropSize,
    naturalSize,
    type CropImageSpec,
    type CropRect,
} from "@/utils/image-crop";

export type PreparedImage = { file: File; previewUrl: string };

type Cropping = { pick: number; img: HTMLImageElement };

// A large image can still be loading when the next one is picked, and the
// dialog can close while Apply is still exporting; only the latest pick may
// reach onCropped.
export default function useImageCrop(
    spec: CropImageSpec,
    fileName: string,
    onCropped: (image: PreparedImage) => void,
) {
    const { t } = useT("settings");
    const [cropping, setCropping] = useState<Cropping | null>(null);
    const pickRef = useRef(0);

    const pick = async (file: File) => {
        const pick = ++pickRef.current;

        try {
            const img = await loadImageFile(file);
            if (pick !== pickRef.current) return;

            if (!fitsCrop(naturalSize(img), spec)) {
                const min = minCropSize(spec);
                toast.error(
                    t("settings:image_dimensions_out_of_range", {
                        minWidth: min.width,
                        minHeight: min.height,
                        max: spec.sourceMaxDimension,
                    }),
                );
                return;
            }
            setCropping({ pick, img });
        } catch {
            if (pick !== pickRef.current) return;
            toast.error(t("settings:image_load_failed"));
        }
    };

    const cancel = () => {
        pickRef.current += 1;
        setCropping(null);
    };

    const apply = async (rect: CropRect) => {
        if (!cropping) return;
        const { pick, img } = cropping;

        try {
            const file = await exportImage(img, rect, spec, fileName);
            const previewUrl = await readFileAsDataUrl(file);
            if (pick !== pickRef.current) return;
            onCropped({ file, previewUrl });
            cancel();
        } catch {
            if (pick !== pickRef.current) return;
            toast.error(t("settings:image_load_failed"));
        }
    };

    return {
        spec,
        image: cropping?.img ?? null,
        dialogKey: cropping?.pick ?? 0,
        pick,
        cancel,
        apply,
    };
}

export type ImageCrop = ReturnType<typeof useImageCrop>;
