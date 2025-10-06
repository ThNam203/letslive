"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import useUser from "../../hooks/user";
import { User } from "../../types/user";
import { GetAllUsers } from "../../lib/api/user";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { cn } from "@/utils/cn";
import {
    HoverCard,
    HoverCardContent,
    HoverCardTrigger,
} from "../ui/hover-card";
import useT from "@/hooks/use-translation";
import { toast } from "react-toastify";

export default function AllChannelsView({
    isMinimized = false,
    minimizeLeftBarIcon,
}: {
    isMinimized?: boolean;
    minimizeLeftBarIcon?: React.ReactNode;
}) {
    const curUser = useUser((state) => state.user);
    const [users, setUsers] = useState<User[]>([]);
    const { t } = useT(["common", "api-response", "fetch-error"]);
    useEffect(() => {
        const fetchAllUsers = async () => {
            await GetAllUsers()
                .then(res => {
                    if (res.success) {
                        setUsers(res.data ?? []);
                    } else {
                        toast(t(`api-response:${res.key}`), {
                            toastId: res.requestId,
                            type: "error",
                        });
                    }
                })
                .catch((_) => {
                    toast(t("fetch-error:client_fetch_error"), { type: "error" });
                });
        };

        fetchAllUsers();
    }, []);

    return (
        <div
            className={cn(
                "flex w-full flex-col items-center gap-2",
                isMinimized ? "" : "px-4",
            )}
        >
            <div className="flex w-full flex-row items-center justify-between">
                {!isMinimized ? (
                    <h2 className="text-xl font-semibold">{t("channels")}</h2>
                ) : null}
                {minimizeLeftBarIcon}
            </div>
            {users.map((user, idx) => {
                if (curUser && curUser.id === user?.id) return null;

                return (
                    <HoverCard key={user.id}>
                        <HoverCardTrigger asChild>
                            <Link
                                href={`/users/${user.id}`}
                                className={cn(
                                    "hover:bg-background-hover flex flex-row items-center gap-2 rounded-full",
                                    isMinimized ? "" : "w-full",
                                )}
                            >
                                <Avatar>
                                    <AvatarImage
                                        src={user.profilePicture}
                                        alt="User avatar"
                                        className="h-10 w-10 rounded-full"
                                    />
                                    <AvatarFallback className="h-10 w-10 rounded-full border border-border">
                                        {user.username.charAt(0).toUpperCase()}
                                    </AvatarFallback>
                                </Avatar>
                                {!isMinimized && (
                                    <span className="text-sm font-semibold">
                                        {user.displayName ?? user.username}
                                    </span>
                                )}
                            </Link>
                        </HoverCardTrigger>
                        <HoverCardContent className="z-10 w-80 bg-muted border-border">
                            <div className="flex gap-4">
                                <Avatar>
                                    <AvatarImage
                                        src={user.profilePicture}
                                        alt="User avatar"
                                        className="h-10 w-10 rounded-full"
                                    />
                                    <AvatarFallback className="h-10 w-10 rounded-full border border-border">
                                        {user.username.charAt(0).toUpperCase()}
                                    </AvatarFallback>
                                </Avatar>
                                <div className="space-y-1">
                                    <h4 className="text-sm font-semibold">
                                        {user.displayName ?? user.username}
                                    </h4>
                                    <p className="text-sm text-muted-foreground">
                                        {user.livestreamInformation
                                            ?.description ?? t("common:no_description")}
                                    </p>
                                    <div className="text-xs text-muted-foreground">
                                        {t("common:followers_with_count", { count: user.followerCount })}
                                    </div>
                                    <div className="text-xs text-muted-foreground">
                                        {t("common:joined")}:{" "}
                                        {new Date(
                                            user.createdAt,
                                        ).toLocaleDateString()}
                                    </div>
                                </div>
                            </div>
                        </HoverCardContent>
                    </HoverCard>
                );
            })}
        </div>
    );
}
