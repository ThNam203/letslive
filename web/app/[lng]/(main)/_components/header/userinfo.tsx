"use client";

import Link from "next/link";

import { useState } from "react";
import { toast } from "@/components/utils/toast";
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
import IconGlobe from "../../../../../components/icons/globe";
import IconPaint from "../../../../../components/icons/paint";
import useT from "@/hooks/use-translation";
import LanguageSwitch from "@/components/utils/language-switch";
import ThemeSwitch from "@/components/utils/theme-switch";
import IconUser from "@/components/icons/user";

export default function UserInfo() {
    const userState = useUser();
    const [isPopoverOpen, setIsPopoverOpen] = useState(false);
    const { t } = useT([
        "auth",
        "common",
        "api-response",
        "fetch-error",
        "settings",
        "theme",
    ]);

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

    return (
        <div className="flex flex-row gap-4">
            <Popover open={isPopoverOpen} onOpenChange={setIsPopoverOpen}>
                <PopoverTrigger>
                    <Avatar className="border-border border">
                        {userState.user ? (
                            <>
                                <AvatarImage
                                    src={userState.user.profilePicture}
                                    alt="avatar"
                                />
                                <AvatarFallback>
                                    {userState.user.username
                                        .charAt(0)
                                        .toUpperCase()}
                                </AvatarFallback>
                            </>
                        ) : (
                            <AvatarFallback>
                                <IconUser className="size-6" />
                            </AvatarFallback>
                        )}
                    </Avatar>
                </PopoverTrigger>
                <PopoverContent className="border-border bg-muted mr-4 w-fit">
                    <div className="flex w-52 flex-col gap-2 rounded-md px-2 pb-2">
                        {userState.user ? (
                            <>
                                <Link
                                    href={`/users/${userState.user.id}`}
                                    className="text-foreground mb-2 w-fit text-lg"
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
                                <div className="flex flex-col gap-2">
                                    <Button asChild>
                                        <Link
                                            href="/settings/profile"
                                            onMouseUp={() =>
                                                setIsPopoverOpen(false)
                                            }
                                            className="flex flex-1 flex-row items-center gap-2"
                                        >
                                            <IconSettings />
                                            <span className="text-primary-foreground text-xs">
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
                                        <span className="text-primary-foreground text-xs">
                                            {t("logout")}
                                        </span>
                                    </Button>
                                </div>
                            </>
                        ) : (
                            <>
                                <div className="flex flex-col gap-3">
                                    <div className="flex flex-col gap-2 pt-2">
                                        <Button asChild className="w-full">
                                            <Link
                                                href="/login"
                                                onMouseUp={() =>
                                                    setIsPopoverOpen(false)
                                                }
                                                className="bg-primary text-primary-foreground hover:bg-primary-hover flex h-8 items-center justify-center rounded-md px-4 text-sm whitespace-nowrap"
                                            >
                                                {t("login")}
                                            </Link>
                                        </Button>
                                        <Button asChild className="w-full">
                                            <Link
                                                href="/signup"
                                                onMouseUp={() =>
                                                    setIsPopoverOpen(false)
                                                }
                                                className="bg-primary text-primary-foreground hover:bg-primary-hover flex h-8 items-center justify-center rounded-md px-4 text-sm whitespace-nowrap"
                                            >
                                                {t("signup")}
                                            </Link>
                                        </Button>
                                    </div>
                                    <div className="flex flex-row items-center justify-between gap-2">
                                        <IconGlobe className="size-6" />
                                        <LanguageSwitch className="h-8 w-full max-w-52" />
                                    </div>
                                    <div className="flex flex-row items-center justify-between gap-2">
                                        <IconPaint className="size-6" />
                                        <ThemeSwitch className="h-8 w-full max-w-52" />
                                    </div>
                                </div>
                            </>
                        )}
                    </div>
                </PopoverContent>
            </Popover>
        </div>
    );
}
