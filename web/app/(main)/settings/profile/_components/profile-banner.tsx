"use client";

import UserAvatar from "@/components/ui/user-avatar";
import useUser from "@/hooks/user";
import { cn } from "@/utils/cn";
import Image from "next/image";
import { useRef } from "react";
import DefaultBackgound from "./default-background";
import ImageHover from "../../_components/image-hover";
import ImageCropDialog from "../../_components/image-crop-dialog";
import useImageCrop, {
    type PreparedImage,
} from "../../_components/use-image-crop";
import useT from "@/hooks/use-translation";
import { toast } from "@/components/utils/toast";
import {
    AVATAR_IMAGE,
    AVATAR_MAX_FILE_MB,
    BACKGROUND_IMAGE,
    BACKGROUND_MAX_FILE_MB,
} from "@/constant/image";
import { readFileAsDataUrl } from "@/utils/file";
import {
    fitsSourceLimit,
    loadImageFile,
    naturalSize,
    exportWholeImage,
} from "@/utils/image-crop";

interface Props {
    className?: string;
    pendingAvatar: PreparedImage | null;
    pendingBackground: PreparedImage | null;
    onAvatarChange: (avatar: PreparedImage | null) => void;
    onBackgroundChange: (background: PreparedImage | null) => void;
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
    const avatarCrop = useImageCrop(AVATAR_IMAGE, "avatar", onAvatarChange);
    // a large image can still be exporting when the next one is picked; only
    // the latest pick may apply
    const backgroundPickRef = useRef(0);

    const handleBackgroundImageChange = async (file: File) => {
        const pick = ++backgroundPickRef.current;

        try {
            const img = await loadImageFile(file);
            if (pick !== backgroundPickRef.current) return;

            if (!fitsSourceLimit(naturalSize(img), BACKGROUND_IMAGE)) {
                toast.error(
                    t("settings:image_too_large_dimensions", {
                        max: BACKGROUND_IMAGE.sourceMaxDimension,
                    }),
                );
                return;
            }

            const exported = await exportWholeImage(
                img,
                BACKGROUND_IMAGE,
                "background",
            );
            const previewUrl = await readFileAsDataUrl(exported);
            if (pick !== backgroundPickRef.current) return;
            onBackgroundChange({ file: exported, previewUrl });
        } catch {
            if (pick !== backgroundPickRef.current) return;
            toast.error(t("settings:image_load_failed"));
        }
    };

    const handleRemoveBackgroundImage = () => {
        backgroundPickRef.current += 1;
        updateUser({ ...user!, backgroundPicture: "" });
        onBackgroundChange(null);
    };

    const handleRemoveProfileImage = () => {
        avatarCrop.cancel();
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
                        onValueChange={avatarCrop.pick}
                        onClick={() => profileImageInputRef.current?.click()}
                        closeIconPosition="bottom"
                        onCloseIconClick={handleRemoveProfileImage}
                        showCloseIcon={Boolean(displayProfilePicture)}
                    />
                </UserAvatar>
            </div>
            <ImageCropDialog
                key={avatarCrop.dialogKey}
                crop={avatarCrop}
                title={t("settings:crop.avatar_title")}
                cropShape="round"
            />
        </div>
    );
}
