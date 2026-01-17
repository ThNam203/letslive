"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import useUser from "@/hooks/user";
import { cn } from "@/utils/cn";
import { useRef } from "react";
import DefaultBackgound from "@/routes/[lng]/(main)/settings/profile/_components/default-background";
import ImageHover from "@/routes/[lng]/(main)/settings/_components/image-hover";

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

    const handleProfileUpdateButtonClick = () => {
        profileImageInputRef.current?.click(); // Trigger file input
    };

    const handleBackgroundUpdateButtonClick = () => {
        backgroundImageInputRef.current?.click(); // Trigger file input
    };

    const handleBackgroundImageChange = (file: File) => {
        updateUser({ ...user!, backgroundPicture: URL.createObjectURL(file) });
        onBackgroundImageChange?.(file);
    };

    const handleProfileImageChange = (file: File) => {
        updateUser({ ...user!, profilePicture: URL.createObjectURL(file) });
        onProfileImageChange?.(file);
    };

    const handleRemoveBackgroundImage = () => {
        updateUser({ ...user!, backgroundPicture: "" });
        onBackgroundImageChange?.(null);
    };

    const handleRemoveProfileImage = () => {
        updateUser({ ...user!, profilePicture: "" });
        onProfileImageChange?.(null);
    };

    return (
        <div className={cn("relative w-full", className)}>
            <div className="relative z-10 h-[300px] w-full overflow-hidden rounded-lg">
                {/* Profile Banner */}
                {user && user.backgroundPicture ? (
                    <img
                        src={
                            user.backgroundPicture ??
                            `https://placehold.co/1200x800/F3F4F6/374151/png?font=playfair-display&text=${
                                user.displayName ?? user.username
                            }`
                        }
                        alt="Profile Banner"
                        className="absolute inset-0 h-full w-full object-cover"
                    />
                ) : (
                    <DefaultBackgound />
                )}
                <ImageHover
                    inputRef={backgroundImageInputRef}
                    onValueChange={handleBackgroundImageChange}
                    onClick={handleBackgroundUpdateButtonClick}
                    onCloseIconClick={handleRemoveBackgroundImage}
                    showCloseIcon={Boolean(user?.backgroundPicture)}
                />
            </div>
            <div className="absolute left-1/2 z-20 -translate-x-1/2 -translate-y-2/3">
                <Avatar className="relative flex h-32 w-32 overflow-hidden rounded-full border-4 border-white">
                    <AvatarImage
                        src={user ? user.profilePicture : ""}
                        alt="user avatar"
                    />
                    <AvatarFallback className="bg-primary text-primary-foreground">
                        {user && user.username[0].toUpperCase()}
                    </AvatarFallback>
                    <ImageHover
                        inputRef={profileImageInputRef}
                        // title={t("settings:profile.update_profile_picture")}
                        onValueChange={handleProfileImageChange}
                        onClick={handleProfileUpdateButtonClick}
                        closeIconPosition="bottom"
                        onCloseIconClick={handleRemoveProfileImage}
                        showCloseIcon={Boolean(user?.profilePicture)}
                    />
                </Avatar>
            </div>
        </div>
    );
}
