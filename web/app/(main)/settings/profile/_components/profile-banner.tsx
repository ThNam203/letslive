"use client";

import UserAvatar from "@/components/ui/user-avatar";
import useUser from "@/hooks/user";
import { cn } from "@/utils/cn";
import Image from "next/image";
import { useRef, useState } from "react";
import DefaultBackgound from "./default-background";
import ImageHover from "../../_components/image-hover";
import useT from "@/hooks/use-translation";
import AvatarCropDialog from "./avatar-crop-dialog";
import { toast } from "@/components/utils/toast";
import {
    AVATAR_ACCEPTED_TYPES,
    AVATAR_SOURCE_MAX_DIMENSION,
    AVATAR_MAX_FILE_MB,
    AVATAR_MIN_DIMENSION,
    AVATAR_UPLOAD_MAX_DIMENSION,
    AVATAR_UPLOAD_QUALITY,
    BACKGROUND_MAX_FILE_MB,
} from "@/constant/image";
import { readFileAsDataUrl } from "@/utils/file";
import {
    exportSquare,
    isWithinDimensions,
    loadImage,
    naturalSize,
    type CropSquare,
} from "@/utils/image-crop";

export type PendingImage = { file: File; previewUrl: string };

type CroppingImage = { pick: number; img: HTMLImageElement };

interface Props {
    className?: string;
    pendingAvatar: PendingImage | null;
    pendingBackground: PendingImage | null;
    onAvatarChange: (avatar: PendingImage | null) => void;
    onBackgroundChange: (background: PendingImage | null) => void;
}

export default function ProfileBanner({
    className,
    pendingAvatar,
    pendingBackground,
    onAvatarChange,
    onBackgroundChange,
}: Props) {
    const { t } = useT(["accessibility", "settings", "api-response"]);
    const user = useUser((state) => state.user);
    const updateUser = useUser((state) => state.updateUser);
    const profileImageInputRef = useRef<HTMLInputElement>(null);
    const backgroundImageInputRef = useRef<HTMLInputElement>(null);
    const [cropping, setCropping] = useState<CroppingImage | null>(null);
    // a large image can still be reading when the next one is picked; only
    // the latest pick may apply
    const avatarPickRef = useRef(0);
    const backgroundPickRef = useRef(0);

    const handleBackgroundImageChange = async (file: File) => {
        const pick = ++backgroundPickRef.current;

        try {
            const previewUrl = await readFileAsDataUrl(file);
            if (pick !== backgroundPickRef.current) return;
            onBackgroundChange({ file, previewUrl });
        } catch {
            if (pick !== backgroundPickRef.current) return;
            toast.error(t("settings:profile.image_load_failed"));
        }
    };

    // a file input fires no change event when the same file is picked again
    const resetProfileImageInput = () => {
        if (profileImageInputRef.current)
            profileImageInputRef.current.value = "";
    };

    const closeCropDialog = () => {
        avatarPickRef.current += 1;
        setCropping(null);
        resetProfileImageInput();
    };

    const handleProfileImageChange = async (file: File) => {
        const pick = ++avatarPickRef.current;

        try {
            const img = await loadImage(await readFileAsDataUrl(file));
            if (pick !== avatarPickRef.current) return;

            if (
                !isWithinDimensions(
                    naturalSize(img),
                    AVATAR_MIN_DIMENSION,
                    AVATAR_SOURCE_MAX_DIMENSION,
                )
            ) {
                toast.error(
                    t("settings:profile.photo_dimensions_out_of_range", {
                        min: AVATAR_MIN_DIMENSION,
                        max: AVATAR_SOURCE_MAX_DIMENSION,
                    }),
                );
                resetProfileImageInput();
                return;
            }
            setCropping({ pick, img });
        } catch {
            if (pick !== avatarPickRef.current) return;
            toast.error(t("settings:profile.image_load_failed"));
            resetProfileImageInput();
        }
    };

    const handleCropApply = async (square: CropSquare) => {
        if (!cropping) return;

        try {
            const file = await exportSquare(cropping.img, square, {
                maxDimension: AVATAR_UPLOAD_MAX_DIMENSION,
                quality: AVATAR_UPLOAD_QUALITY,
                fileName: "avatar",
            });
            onAvatarChange({ file, previewUrl: await readFileAsDataUrl(file) });
            closeCropDialog();
        } catch {
            toast.error(t("settings:profile.image_load_failed"));
        }
    };

    const handleRemoveBackgroundImage = () => {
        backgroundPickRef.current += 1;
        updateUser({ ...user!, backgroundPicture: "" });
        onBackgroundChange(null);
    };

    const handleRemoveProfileImage = () => {
        avatarPickRef.current += 1;
        updateUser({ ...user!, profilePicture: "" });
        onAvatarChange(null);
    };

    const displayBackground =
        pendingBackground?.previewUrl ?? user?.backgroundPicture;
    const displayProfilePicture =
        pendingAvatar?.previewUrl ?? user?.profilePicture;

    return (
        <div className={cn("relative w-full", className)}>
            <div className="relative z-10 h-[300px] w-full overflow-hidden rounded-lg">
                {/* Profile Banner */}
                {displayBackground ? (
                    <Image
                        src={displayBackground}
                        alt={t("profile_banner")}
                        fill={true}
                        className="object-cover"
                        loading="eager"
                        unoptimized
                    />
                ) : (
                    <DefaultBackgound />
                )}
                <ImageHover
                    maxFileMB={BACKGROUND_MAX_FILE_MB}
                    inputRef={backgroundImageInputRef}
                    onValueChange={handleBackgroundImageChange}
                    onClick={() => backgroundImageInputRef.current?.click()}
                    onCloseIconClick={handleRemoveBackgroundImage}
                    showCloseIcon={Boolean(displayBackground)}
                />
            </div>
            <div className="absolute left-1/2 z-20 -translate-x-1/2 -translate-y-2/3">
                <UserAvatar
                    size="lg"
                    src={displayProfilePicture}
                    name={user?.username}
                    alt={t("user_avatar")}
                    className="border-4 border-white"
                    fallbackClassName="bg-primary text-primary-foreground"
                >
                    <ImageHover
                        maxFileMB={AVATAR_MAX_FILE_MB}
                        inputRef={profileImageInputRef}
                        onValueChange={handleProfileImageChange}
                        onClick={() => profileImageInputRef.current?.click()}
                        closeIconPosition="bottom"
                        onCloseIconClick={handleRemoveProfileImage}
                        showCloseIcon={Boolean(displayProfilePicture)}
                        accept={AVATAR_ACCEPTED_TYPES}
                    />
                </UserAvatar>
            </div>
            <AvatarCropDialog
                key={cropping?.pick ?? 0}
                image={cropping?.img ?? null}
                onCancel={closeCropDialog}
                onApply={handleCropApply}
            />
        </div>
    );
}
