"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import useUser from "@/hooks/user";
import { cn } from "@/utils/cn";
import Image from "next/image";
import { useEffect, useRef, useState } from "react";
import DefaultBackgound from "./default-background";
import ImageHover from "../../_components/image-hover";

interface Props {
    className?: string;
    onProfileImageChange?: (file: File | null) => void;
    onBackgroundImageChange?: (file: File | null) => void;
}
export default function ProfileBanner({
    className,
    onBackgroundImageChange,
    onProfileImageChange,
}: Props) {
    const user = useUser((state) => state.user);
    const updateUser = useUser((state) => state.updateUser);
    const profileImageInputRef = useRef<HTMLInputElement>(null);
    const backgroundImageInputRef = useRef<HTMLInputElement>(null);

    const [previewProfileUrl, setPreviewProfileUrl] = useState<string | null>(null);
    const [previewBackgroundUrl, setPreviewBackgroundUrl] = useState<string | null>(null);
    const previewProfileUrlRef = useRef<string | null>(null);
    const previewBackgroundUrlRef = useRef<string | null>(null);

    // Revoke blob URLs on unmount
    useEffect(() => {
        return () => {
            if (previewProfileUrlRef.current) URL.revokeObjectURL(previewProfileUrlRef.current);
            if (previewBackgroundUrlRef.current) URL.revokeObjectURL(previewBackgroundUrlRef.current);
        };
    }, []);

    // When the store is updated with a real server URL after save, clear the local preview
    useEffect(() => {
        if (!user?.profilePicture?.startsWith("blob:")) {
            if (previewProfileUrlRef.current) {
                URL.revokeObjectURL(previewProfileUrlRef.current);
                previewProfileUrlRef.current = null;
            }
            setPreviewProfileUrl(null);
        }
    }, [user?.profilePicture]);

    useEffect(() => {
        if (!user?.backgroundPicture?.startsWith("blob:")) {
            if (previewBackgroundUrlRef.current) {
                URL.revokeObjectURL(previewBackgroundUrlRef.current);
                previewBackgroundUrlRef.current = null;
            }
            setPreviewBackgroundUrl(null);
        }
    }, [user?.backgroundPicture]);

    const handleProfileUpdateButtonClick = () => {
        profileImageInputRef.current?.click();
    };

    const handleBackgroundUpdateButtonClick = () => {
        backgroundImageInputRef.current?.click();
    };

    const handleBackgroundImageChange = (file: File) => {
        if (previewBackgroundUrlRef.current) URL.revokeObjectURL(previewBackgroundUrlRef.current);
        const blobUrl = URL.createObjectURL(file);
        previewBackgroundUrlRef.current = blobUrl;
        setPreviewBackgroundUrl(blobUrl);
        onBackgroundImageChange?.(file);
    };

    const handleProfileImageChange = (file: File) => {
        if (previewProfileUrlRef.current) URL.revokeObjectURL(previewProfileUrlRef.current);
        const blobUrl = URL.createObjectURL(file);
        previewProfileUrlRef.current = blobUrl;
        setPreviewProfileUrl(blobUrl);
        onProfileImageChange?.(file);
    };

    const handleRemoveBackgroundImage = () => {
        if (previewBackgroundUrlRef.current) {
            URL.revokeObjectURL(previewBackgroundUrlRef.current);
            previewBackgroundUrlRef.current = null;
        }
        setPreviewBackgroundUrl(null);
        updateUser({ ...user!, backgroundPicture: "" });
        onBackgroundImageChange?.(null);
    };

    const handleRemoveProfileImage = () => {
        if (previewProfileUrlRef.current) {
            URL.revokeObjectURL(previewProfileUrlRef.current);
            previewProfileUrlRef.current = null;
        }
        setPreviewProfileUrl(null);
        updateUser({ ...user!, profilePicture: "" });
        onProfileImageChange?.(null);
    };

    const displayBackground = previewBackgroundUrl ?? user?.backgroundPicture;
    const displayProfilePicture = previewProfileUrl ?? (user ? user.profilePicture : "");

    return (
        <div className={cn("relative w-full", className)}>
            <div className="relative z-10 h-[300px] w-full overflow-hidden rounded-lg">
                {/* Profile Banner */}
                {displayBackground ? (
                    <Image
                        src={displayBackground}
                        alt="Profile Banner"
                        fill={true}
                        className="object-cover"
                        unoptimized
                    />
                ) : (
                    <DefaultBackgound />
                )}
                <ImageHover
                    inputRef={backgroundImageInputRef}
                    onValueChange={handleBackgroundImageChange}
                    onClick={handleBackgroundUpdateButtonClick}
                    onCloseIconClick={handleRemoveBackgroundImage}
                    showCloseIcon={Boolean(displayBackground)}
                />
            </div>
            <div className="absolute left-1/2 z-20 -translate-x-1/2 -translate-y-2/3">
                <Avatar className="relative flex h-32 w-32 overflow-hidden rounded-full border-4 border-white">
                    <AvatarImage
                        src={displayProfilePicture}
                        alt="user avatar"
                    />
                    <AvatarFallback className="bg-primary text-primary-foreground">
                        {user && user.username[0].toUpperCase()}
                    </AvatarFallback>
                    <ImageHover
                        inputRef={profileImageInputRef}
                        onValueChange={handleProfileImageChange}
                        onClick={handleProfileUpdateButtonClick}
                        closeIconPosition="bottom"
                        onCloseIconClick={handleRemoveProfileImage}
                        showCloseIcon={Boolean(displayProfilePicture)}
                    />
                </Avatar>
            </div>
        </div>
    );
}
