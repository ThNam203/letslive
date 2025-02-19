"use client";

import {
    LuBell,
    LuLogOut,
    LuMessageSquare,
    LuSettings
} from "react-icons/lu";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover";
import useUser from "@/hooks/user";
import { useEffect } from "react";
import { FetchError } from "@/types/fetch-error";
import { toast } from "react-toastify";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { Logout } from "@/lib/api/auth";

export default function UserInfo() {
    const userState = useUser();
    const router = useRouter();

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
                if (e instanceof FetchError && e.isClientError) {
                    toast(e.message, {
                        toastId: "fetch-user-error",
                        type: "error",
                    });
                    router.push("/login");
                } else {
                    toast("An unknown error occurred", {
                        toastId: "fetch-user-error",
                        type: "error",
                    });
                }
            });
        };

        fetchUser();
    }, []);
    return userState.user ? (
        <div className="flex flex-row gap-4">
            <Button>
                <LuBell size={16} />
            </Button>
            <Button>
                <LuMessageSquare size={16} />
            </Button>

            <Popover>
                <PopoverTrigger>
                    <Image
                        src={
                            userState.user.profilePicture ??
                            "https://github.com/shadcn.png"
                        }
                        alt="avatar"
                        width={32}
                        height={32}
                        className="rounded-full"
                    />
                </PopoverTrigger>
                <PopoverContent className="w-60 mr-4">
                    <div className="pb-2 px-2 bg-white rounded-md shadow-primaryShadow flex flex-col gap-2">
                        <Link href={`/users/${userState.user.id}`}>
                            <span className="text-lg text-gray-900">
                                {userState.user.displayName ??
                                    userState.user.username}
                            </span>
                        </Link>
                        <Button>
                            <Link href="/settings/profile">
                                <div className="flex flex-row gap-2 items-center">
                                    <LuSettings />
                                    <span className="text-xs">Setting</span>
                                </div>
                            </Link>
                        </Button>
                        <Button onClick={logoutHandler}>
                            <div className="flex flex-row gap-2 items-center text-red-500">
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
                className="whitespace-nowrap bg-white border-1 rounded-md hover:bg-gray-200 text-gray-900 border-gray-700"
            >
                Log In
            </Link>
            <Link
                className="whitespace-nowrap bg-white border-1 rounded-md hover:bg-gray-200 text-gray-900 border-gray-700"
                href="/signup"
            >
                Sign Up
            </Link>
        </div>
    );
}
