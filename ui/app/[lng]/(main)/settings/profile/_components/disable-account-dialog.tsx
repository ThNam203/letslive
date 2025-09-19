"use client";

import IconLoader from "@/components/icons/loader";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import useUser from "@/hooks/user";
import { Logout } from "@/lib/api/auth";
import { UpdateProfile } from "@/lib/api/user";
import { UserStatus } from "@/types/user";
import { useParams, useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "react-toastify";
import useT from "@/hooks/use-translation";

export default function DisableAccountDialog({
    isUpdatingProfile,
}: {
    isUpdatingProfile: boolean;
}) {
    const user = useUser((state) => state.user);
    const clearUser = useUser((state) => state.clearUser);
    const [isDisablingAccount, setIsDisablingAccount] = useState(false);
    const [isOpen, setIsOpen] = useState(false);
    const { t } = useT("settings");

    const logoutHandler = async () => {
        const { fetchError } = await Logout();
        if (fetchError) {
            toast(fetchError.message, {
                toastId: "logout-error",
                type: "error",
            });
        } else {
            clearUser();
        }
    };

    const handleDisableAccount = async () => {
        try {
            setIsDisablingAccount(true);
            const { fetchError } = await UpdateProfile({
                id: user!.id,
                status: UserStatus.DISABLED,
            }).finally(() => setIsDisablingAccount(false));
            if (fetchError) {
                toast.error(fetchError.message);
                return;
            }
            await logoutHandler();
        } catch (error) {
            toast.error(t("settings:disable.unknown_error"));
        } finally {
            setIsOpen(false);
            setIsDisablingAccount(false);
        }
    };
    return (
        <Dialog open={isOpen} onOpenChange={setIsOpen}>
            <DialogTrigger asChild>
                <button
                    disabled={isUpdatingProfile || isDisablingAccount}
                    className="rounded-md bg-destructive px-4 py-2 text-sm font-medium text-destructive-foreground hover:bg-destructive-hover"
                >
                    {t("settings:disable.button")}
                </button>
            </DialogTrigger>
            <DialogContent className="bg-background text-foreground">
                <DialogHeader>
                    <DialogTitle>{t("settings:disable.dialog.title")}</DialogTitle>
                    <DialogDescription>
                        {t("settings:disable.dialog.description")}
                    </DialogDescription>
                </DialogHeader>

                <DialogFooter>
                    <DialogClose asChild>
                        <Button variant="outline">{t("common:cancel")}</Button>
                    </DialogClose>
                    <Button
                        disabled={isUpdatingProfile || isDisablingAccount}
                        onClick={handleDisableAccount}
                    >
                        {t("settings:disable.dialog.confirm")}
                        {isDisablingAccount && <IconLoader className="ml-1" />}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}

