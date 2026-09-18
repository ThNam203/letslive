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
import { useLogout } from "@/hooks/queries/use-auth-mutations";
import { useUpdateProfile } from "@/hooks/queries/use-profile-mutations";
import { UserStatus } from "@/types/user";
import { useState } from "react";
import useT from "@/hooks/use-translation";

export default function DisableAccountDialog({
    isUpdatingProfile,
}: {
    isUpdatingProfile: boolean;
}) {
    const updateProfile = useUpdateProfile();
    const logout = useLogout();
    const [isOpen, setIsOpen] = useState(false);
    const { t } = useT(["settings", "api-response", "fetch-error"]);

    const handleDisableAccount = () => {
        updateProfile.mutate(
            { status: UserStatus.DISABLED },
            {
                onSuccess: () => logout.mutate(),
                onSettled: () => setIsOpen(false),
            },
        );
    };

    return (
        <Dialog open={isOpen} onOpenChange={setIsOpen}>
            <DialogTrigger asChild>
                <button
                    disabled={isUpdatingProfile || updateProfile.isPending}
                    className="bg-destructive text-destructive-foreground hover:bg-destructive-hover rounded-md px-4 py-2 text-sm font-medium"
                >
                    {t("settings:disable.button")}
                </button>
            </DialogTrigger>
            <DialogContent className="bg-background text-foreground">
                <DialogHeader>
                    <DialogTitle>
                        {t("settings:disable.dialog.title")}
                    </DialogTitle>
                    <DialogDescription>
                        {t("settings:disable.dialog.description")}
                    </DialogDescription>
                </DialogHeader>

                <DialogFooter>
                    <DialogClose asChild>
                        <Button variant="outline">{t("common:cancel")}</Button>
                    </DialogClose>
                    <Button
                        disabled={isUpdatingProfile || updateProfile.isPending}
                        onClick={handleDisableAccount}
                    >
                        {t("settings:disable.dialog.confirm")}
                        {updateProfile.isPending && (
                            <IconLoader className="ml-1" />
                        )}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
