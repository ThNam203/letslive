"use client";

import Image from "next/image";
import { User } from "../../../../types/user";
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "../../../../components/ui/avatar";

export default function ProfileHeader({
  user,
}: {
  user: User;
  updateUser: (newUserInfo: User) => void;
}) {
  return (
    <div className="relative">
      <div className="relative w-full h-[300px] bg-gray-100 rounded-md overflow-hidden shadow">
        {/* Profile Banner */}
        <Image
          src={
            user.backgroundPicture ??
            `https://placehold.co/1200x800/F3F4F6/374151/png?font=playfair-display&text=${
              user.displayName ?? user.username
            }`
          }
          alt="Profile Banner"
          layout="fill"
          objectFit="cover"
        />
      </div>
      <div className="sm:px-6 px-4">
        <div className="relative -mt-16 sm:-mt-24">
          <div className="absolute">
            <div className="flex flex-row items-end gap-4">
              <Avatar className="rounded-full border-4 border-white w-32 h-32">
                <AvatarImage src={user.profilePicture} alt="user avatar" />
                <AvatarFallback>
                  {user.username[0].toUpperCase()}
                </AvatarFallback>
              </Avatar>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
