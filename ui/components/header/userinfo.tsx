"use client";

import Link from "next/link";

import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import { useRouter } from "next/navigation";
import Image from "next/image";
import useUser from "../../hooks/user";
import { Logout } from "../../lib/api/auth";
import { FetchError } from "../../types/fetch-error";
import { Button } from "../ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "../ui/popover";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import IconSettings from "../icons/settings";
import IconLogOut from "../icons/log-out";

export default function UserInfo() {
  const userState = useUser();
  const router = useRouter();
  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  const logoutHandler = async () => {
    const { fetchError } = await Logout();
    if (fetchError) {
      toast(fetchError.message, {
        toastId: "logout-error",
        type: "error",
      });
    } else {
      userState.clearUser();
      router.push("/login");
    }
  };

  useEffect(() => {
    const fetchUser = async () => {
      userState.fetchUser().catch((e) => {
        // if (e instanceof FetchError && e.isClientError) {
        //     toast(e.message, {
        //         toastId: "fetch-user-error",
        //         type: "error",
        //     });
        //     router.push("/login");
        // } else {
        //     toast("An unknown error occurred", {
        //         toastId: "fetch-user-error",
        //         type: "error",
        //     });
        // }
      });
    };

    fetchUser();
  }, []);
  return userState.user ? (
    <div className="flex flex-row gap-4">
      {/* <Button>
                <LuBell size={16} />
            </Button>
            <Button>
                <LuMessageSquare size={16} />
            </Button> */}

      <Popover open={isPopoverOpen} onOpenChange={setIsPopoverOpen}>
        <PopoverTrigger>
          <Avatar>
            <AvatarImage src={userState.user.profilePicture} alt="avatar" />
            <AvatarFallback>
              {userState.user.username.charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
        </PopoverTrigger>
        <PopoverContent className="w-100 mr-4">
          <div className="pb-2 px-2 bg-white rounded-md shadow-primaryShadow flex flex-col gap-2">
            <Link
              href={`/users/${userState.user.id}`}
              className="text-lg text-gray-900 w-fit mb-2"
              onMouseUp={() => setIsPopoverOpen(false)}
            >
              <p>{userState.user.displayName ?? userState.user.username}</p>
              <p className="text-sm">@{userState.user.username}</p>
            </Link>
            <div className="flex gap-2">
              <Button asChild>
                <Link
                  href="/settings/profile"
                  onMouseUp={() => setIsPopoverOpen(false)}
                  className="flex flex-1 flex-row gap-2 items-center"
                >
                  <IconSettings />
                  <span className="text-xs">Setting</span>
                </Link>
              </Button>
              <Button
                onClick={logoutHandler}
                className="flex flex-1 flex-row gap-2 items-center text-red-500 hover:cursor-pointer"
              >
                <IconLogOut />
                <span className="text-xs">Log Out</span>
              </Button>
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  ) : (
    <div className="flex flex-row gap-2">
      <Link
        href="/login"
        className="whitespace-nowrap bg-primary border-1 rounded-md hover:bg-primary-hover text-primary-foreground border-border text-sm py-1 px-4"
      >
        Log in
      </Link>
      <Link
        className="whitespace-nowrap bg-primary border-1 rounded-md hover:bg-primary-hover text-primary-foreground border-border text-sm py-1 px-4"
        href="/signup"
      >
        Sign up
      </Link>
    </div>
  );
}
