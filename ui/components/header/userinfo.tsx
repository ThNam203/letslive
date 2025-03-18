"use client";

import { LuBell, LuLogOut, LuMessageSquare, LuSettings } from "react-icons/lu";
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
                        <AvatarImage
                            src={userState.user.profilePicture ?? "https://github.com/shadcn.png"}
                            alt="avatar"
                        />
                        <AvatarFallback>
                            {userState.user.username.charAt(0).toUpperCase()}
                        </AvatarFallback>
                    </Avatar>
                </PopoverTrigger>
                <PopoverContent className="w-60 mr-4">
                    <div className="pb-2 px-2 bg-white rounded-md shadow-primaryShadow flex flex-col gap-2">
                        <Link
                            href={`/users/${userState.user.id}`}
                            className="text-lg text-gray-900 w-fit mb-2"
                            onMouseUp={() => setIsPopoverOpen(false)}
                        >
                            <p>
                                {userState.user.displayName ??
                                    userState.user.username}
                            </p>
                            <p className="text-sm">@{userState.user.username}</p>
                        </Link>
                        <Button asChild>
                            <Link href="/settings/profile" onMouseUp={() => setIsPopoverOpen(false)}>
                                <div className="flex flex-row gap-2 items-center">
                                    <LuSettings />
                                    <span className="text-xs">Setting</span>
                                </div>
                            </Link>
                        </Button>
                        <Button asChild>
                            <div
                                onClick={logoutHandler}
                                className="flex flex-row gap-2 items-center text-red-500 hover:cursor-pointer"
                            >
                                <LuLogOut />
                                <span className="text-xs">Log Out</span>
                            </div>
                        </Button>
                    </div>
                </PopoverContent>
            </Popover>
        </div>
    ) : (
        <div className="flex flex-row gap-2">
            <Link
                href="/login"
                className="whitespace-nowrap bg-white border-1 rounded-md hover:bg-gray-200 text-gray-900 border-gray-700 text-sm py-1 px-4"
            >
                Log in
            </Link>
            <Link
                className="whitespace-nowrap bg-white border-1 rounded-md hover:bg-gray-200 text-gray-900 border-gray-700 text-sm py-1 px-4"
                href="/signup"
            >
                Sign up
            </Link>
        </div>
    );
}
