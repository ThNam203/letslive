"use client";

import Link from "next/link";

import { useState } from "react";
import { toast } from "react-toastify";
import useUser from "../../hooks/user";
import { Logout } from "../../lib/api/auth";
import { Button } from "../ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "../ui/popover";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import IconSettings from "../icons/settings";
import IconLogOut from "../icons/log-out";
import useT from "@/hooks/use-translation";

export default function UserInfo() {
  const userState = useUser();
  const [isPopoverOpen, setIsPopoverOpen] = useState(false);
  const { t } = useT(["auth", "common", "api-response", "fetch-error"]);

  const logoutHandler = async () => {
    await Logout()
      .then(res => {
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
            <AvatarImage src={userState.user.profilePicture} alt="avatar" />
            <AvatarFallback>
              {userState.user.username.charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
        </PopoverTrigger>
        <PopoverContent className="w-100 mr-4 bg-muted border-border">
          <div className="pb-2 px-2 rounded-md flex flex-col gap-2">
            <Link
              href={`/users/${userState.user.id}`}
              className="text-lg text-foreground w-fit mb-2"
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
                  <span className="text-xs text-primary-foreground">{t("common:setting")}</span>
                </Link>
              </Button>
              <Button
                onClick={logoutHandler}
                className="flex flex-1 flex-row gap-2 items-center hover:cursor-pointer"
                variant={"destructive"}
              >
                <IconLogOut />
                <span className="text-xs text-primary-foreground">{t("logout")}</span>
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
        className="whitespace-nowrap bg-primary border rounded-md hover:bg-primary-hover text-primary-foreground border-border text-sm py-1 px-4"
      >
        {t("login")}
      </Link>
      <Link
        className="whitespace-nowrap bg-primary border rounded-md hover:bg-primary-hover text-primary-foreground border-border text-sm py-1 px-4"
        href="/signup"
      >
        {t("signup")}
      </Link>
    </div>
  );
}
