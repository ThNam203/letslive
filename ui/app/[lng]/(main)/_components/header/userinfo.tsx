"use client";

import Link from "next/link";

import { useState } from "react";
import { toast } from "react-toastify";
import useUser from "../../../../../hooks/user";
import { Logout } from "../../../../../lib/api/auth";
import { Button } from "../../../../../components/ui/button";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "../../../../../components/ui/popover";
import {
    Avatar,
    AvatarFallback,
    AvatarImage,
} from "../../../../../components/ui/avatar";
import IconSettings from "../../../../../components/icons/settings";
import IconLogOut from "../../../../../components/icons/log-out";
import useT from "@/hooks/use-translation";

export default function UserInfo() {
    const userState = useUser();
    const [isPopoverOpen, setIsPopoverOpen] = useState(false);
    const { t } = useT(["auth", "common", "api-response", "fetch-error"]);

    const logoutHandler = async () => {
        await Logout()
            .then((res) => {
                if (res.statusCode === 204) {
                    userState.clearUser();
                } else {
                    toast(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                        type: "error",
                    });
                }
            })
            .catch((err) => {
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "client-fetch-error-id",
                    type: "error",
                });
            });
    };

    return userState.user ? (
        <div className="flex flex-row gap-4">
            <Popover open={isPopoverOpen} onOpenChange={setIsPopoverOpen}>
                <PopoverTrigger>
                    <Avatar className="border border-border">
                        <AvatarImage
                            src={userState.user.profilePicture}
                            alt="avatar"
                        />
                        <AvatarFallback>
                            {userState.user.username.charAt(0).toUpperCase()}
                        </AvatarFallback>
                    </Avatar>
                </PopoverTrigger>
                <PopoverContent className="w-100 mr-4 border-border bg-muted">
                    <div className="flex flex-col gap-2 rounded-md px-2 pb-2">
                        <Link
                            href={`/users/${userState.user.id}`}
                            className="mb-2 w-fit text-lg text-foreground"
                            onMouseUp={() => setIsPopoverOpen(false)}
                        >
                            <p>
                                {userState.user.displayName ??
                                    userState.user.username}
                            </p>
                            <p className="text-sm">
                                @{userState.user.username}
                            </p>
                        </Link>
                        <div className="flex gap-2">
                            <Button asChild>
                                <Link
                                    href="/settings/profile"
                                    onMouseUp={() => setIsPopoverOpen(false)}
                                    className="flex flex-1 flex-row items-center gap-2"
                                >
                                    <IconSettings />
                                    <span className="text-xs text-primary-foreground">
                                        {t("common:setting")}
                                    </span>
                                </Link>
                            </Button>
                            <Button
                                onClick={logoutHandler}
                                className="flex flex-1 flex-row items-center gap-2 hover:cursor-pointer"
                                variant={"destructive"}
                            >
                                <IconLogOut />
                                <span className="text-xs text-primary-foreground">
                                    {t("logout")}
                                </span>
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
                className="flex h-8 items-center whitespace-nowrap rounded-md bg-primary px-4 text-sm text-primary-foreground hover:bg-primary-hover"
            >
                {t("login")}
            </Link>
            <Link
                className="flex h-8 items-center whitespace-nowrap rounded-md bg-primary px-4 text-sm text-primary-foreground hover:bg-primary-hover"
                href="/signup"
            >
                {t("signup")}
            </Link>
        </div>
    );
}
