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
import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "react-toastify";

export default function DisableAccountDialog({
    isUpdatingProfile,
}: {
    isUpdatingProfile: boolean;
}) {
    const user = useUser((state) => state.user);
    const clearUser = useUser((state) => state.clearUser);
    const router = useRouter();
    const [isDisablingAccount, setIsDisablingAccount] = useState(false);
    const [isOpen, setIsOpen] = useState(false);

    const logoutHandler = async () => {
        const { fetchError } = await Logout();
        if (fetchError) {
            toast(fetchError.message, {
                toastId: "logout-error",
                type: "error",
            });
        } else {
            clearUser();
            router.push("/login");
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
            // await logoutHandler();
        } catch (error) {
            toast.error("An unknown error occurred");
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
                    Disable Your Let&apos;s Live Account
                </button>
            </DialogTrigger>
            <DialogContent className="bg-background text-foreground">
                <DialogHeader>
                    <DialogTitle>Disable your account!</DialogTitle>
                    <DialogDescription>
                        Are you sure you want to disable your account? You can
                        reactivate your account at any time by logging back in.
                    </DialogDescription>
                </DialogHeader>

                <DialogFooter>
                    <DialogClose asChild>
                        <Button variant="outline">Cancel</Button>
                    </DialogClose>
                    <Button
                        disabled={isUpdatingProfile || isDisablingAccount}
                        onClick={handleDisableAccount}
                    >
                        I understand!
                        {isDisablingAccount && <IconLoader className="ml-1" />}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
