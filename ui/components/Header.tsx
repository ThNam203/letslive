"use client";
import {
  LuBell,
  LuCopy,
  LuHeart,
  LuHome,
  LuLogOut,
  LuMessageSquare,
  LuMoreVertical,
  LuPodcast,
  LuSettings,
  LuUser as UserUI,
} from "react-icons/lu";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { use, useEffect, useState } from "react";
import { User } from "@/types/user";
import IconButton from "@/components/buttons/IconBtn";
import { SearchInput } from "@/components/Input";
import TextButton from "@/components/buttons/TextButton";
import Image from "next/image";
import Separator from "@/components/Separator";
import { DefaultOption } from "@/components/Option";
import { Popover, PopoverContent, PopoverTrigger } from "@nextui-org/popover";

export const Header = () => {
  const router = useRouter();
  const [showPopover, setShowPopover] = useState(false);
  const [thisUser, setThisUser] = useState<User | null>(null);

  return (
    <nav className="w-full h-12 flex flex-row items-center justify-between text-xl font-semibold text-primaryWord bg-white px-4 py-2 shadow z-[49]">
      <div className="flex flex-row md:gap-10 max-md:gap-4 items-center">
        <Link href="/" className="hover:text-primary">
          <span className="max-md:hidden">Home</span>
          <LuHome size={20} className="md:hidden" />
        </Link>
        <Link href="/following" className="hover:text-primary">
          <span className="max-md:hidden">Following</span>
          <LuHeart size={20} className="md:hidden" />
        </Link>
        <Link href="/browse" className="hover:text-primary">
          <span className="max-md:hidden">Browse</span>
          <LuCopy size={20} className="md:hidden" />
        </Link>
        <IconButton icon={<LuMoreVertical />} />
      </div>

      <div className="lg:w-[400px] max-lg:w mx-2">
        <SearchInput
          id="search-input"
          placeholder="Search (Not implemented)"
          className="text-base w-full"
        />
      </div>

      {thisUser ? (
        <div className="flex flex-row gap-4">
          <TextButton
            content="Stream now"
            iconAfter={<LuPodcast size={16} />}
            className="bg-primary hover:bg-secondary text-white"
            onClick={() => {
              router.push("/livestreaming");
            }}
          />
          <IconButton icon={<LuBell size={16} />} />
          <IconButton icon={<LuMessageSquare size={16} />} />

          <Popover
            isOpen={showPopover}
            onOpenChange={setShowPopover}
            placement="bottom-end"
            showArrow={true}
          >
            <PopoverTrigger>
              <button
                className="bg-[#69ffc3]"
                >
                <UserUI size={16} strokeWidth={3} />
              </button>
            </PopoverTrigger>
            <PopoverContent>
              <div
                className="py-4 px-2 bg-white rounded-md shadow-primaryShadow flex flex-col"
                onClick={() => setShowPopover(false)}
              >
                <div className="flex flex-row gap-2 items-center">
                  <button
                    className="bg-[#69ffc3] w-8 h-8"
                    >
                    <UserUI size={16} strokeWidth={3} />
                    </button>
                  <span className="text-xs font-semibold">
                    {thisUser.username}
                  </span>
                </div>

                <Separator classname="my-2" />
                <DefaultOption
                  content={
                    <div className="flex flex-row gap-2 items-center">
                      <LuSettings />
                      <span className="text-xs">Setting</span>
                    </div>
                  }
                  onClick={() => {
                    router.push("/setting");
                  }}
                />

                <Separator classname="my-2" />
                <DefaultOption
                  content={
                    <div className="flex flex-row gap-2 items-center text-red-500">
                      <LuLogOut/>
                      <span className="text-xs">Log Out</span>
                    </div>
                  }
                                  />
              </div>
            </PopoverContent>
          </Popover>
        </div>
      ) : (
        <div className="flex flex-row gap-2">
          <TextButton
            content="Log In"
            onClick={() => {
              router.push("/login");
            }}
            className="whitespace-nowrap"
          />
          <TextButton
            content="Sign Up"
            className="text-white bg-primary hover:bg-primary/80 whitespace-nowrap"
            onClick={() => {
              router.push("/signup");
            }}
          />
        </div>
      )}
    </nav>
  );
};