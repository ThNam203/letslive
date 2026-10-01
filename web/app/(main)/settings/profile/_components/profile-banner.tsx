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
import { AVATAR_ACCEPTED_TYPES } from "@/constant/image";
import { readFileAsDataUrl } from "@/utils/file";
import {
    exportAvatar,
    isAllowedAvatarSize,
    loadImage,
    naturalSize,
    type CropSquare,
} from "@/utils/avatar-crop";

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
    // a large image can still be loading when the next one is picked; only
    // the latest pick may open the dialog
    const avatarPickRef = useRef(0);

    const handleBackgroundImageChange = async (file: File) => {
        try {
            onBackgroundChange({
                file,
                previewUrl: await readFileAsDataUrl(file),
            });
        } catch {
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

            if (!isAllowedAvatarSize(naturalSize(img))) {
                toast.error(
                    t("api-response:res_err_image_dimensions_out_of_range"),
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
            const file = await exportAvatar(cropping.img, square);
            onAvatarChange({ file, previewUrl: await readFileAsDataUrl(file) });
            closeCropDialog();
        } catch {
            toast.error(t("settings:profile.image_load_failed"));
        }
    };

    const handleRemoveBackgroundImage = () => {
        updateUser({ ...user!, backgroundPicture: "" });
        onBackgroundChange(null);
    };

    const handleRemoveProfileImage = () => {
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
