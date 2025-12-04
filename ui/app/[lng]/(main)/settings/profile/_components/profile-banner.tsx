"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import useUser from "@/hooks/user";
import { cn } from "@/utils/cn";
import Image from "next/image";
import { useRef } from "react";
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
    <div className={cn("w-full relative", className)}>
      <div className="z-10 relative w-full h-[300px] rounded-lg overflow-hidden">
        {/* Profile Banner */}
        {user && user.backgroundPicture ? (
          <Image
            src={
              user.backgroundPicture ??
              `https://placehold.co/1200x800/F3F4F6/374151/png?font=playfair-display&text=${
                user.displayName ?? user.username
              }`
            }
            alt="Profile Banner"
            fill={true}
            className="object-cover"
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
      <div className="z-20 absolute -translate-y-2/3 left-1/2 -translate-x-1/2">
        <Avatar className="relative flex rounded-full border-4 border-white w-32 h-32 overflow-hidden">
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
