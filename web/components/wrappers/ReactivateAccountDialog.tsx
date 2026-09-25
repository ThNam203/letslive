"use client";

import IconLoader from "@/components/icons/loader";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import useUser from "@/hooks/user";
import { Logout } from "@/lib/api/auth";
import { UpdateProfile } from "@/lib/api/user";
import { UserStatus } from "@/types/user";
import { useState } from "react";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";

export default function ReactivateAccountDialog() {
    const user = useUser((state) => state.user);
    const updateUser = useUser((state) => state.updateUser);
    const clearUser = useUser((state) => state.clearUser);
    const [isReactivating, setIsReactivating] = useState(false);
    const [isLoggingOut, setIsLoggingOut] = useState(false);
    const { t } = useT(["settings", "api-response", "fetch-error"]);

    const isOpen = user?.status === UserStatus.DISABLED;

    const handleReactivate = async () => {
        setIsReactivating(true);
        await UpdateProfile({
            status: UserStatus.NORMAL,
        })
            .then((res) => {
                if (res.success) {
                    updateUser({ status: UserStatus.NORMAL });
                    toast(t("settings:reactivate.success"), {
                        type: "success",
                    });
                } else {
                    toast(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                        type: "error",
                    });
                }
            })
            .catch((_) => {
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "client-fetch-error-id",
                    type: "error",
                });
            })
            .finally(() => setIsReactivating(false));
    };

    const handleDecline = async () => {
        setIsLoggingOut(true);
        await Logout()
            .then((res) => {
                if (res.statusCode === 204) {
                    clearUser();
                } else {
                    toast(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                        type: "error",
                    });
                }
            })
            .catch((_) => {
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "client-fetch-error-id",
                    type: "error",
                });
            })
            .finally(() => setIsLoggingOut(false));
    };

    return (
        <Dialog open={isOpen}>
            <DialogContent
                className="bg-background text-foreground"
                showCloseButton={false}
                onInteractOutside={(e) => e.preventDefault()}
                onEscapeKeyDown={(e) => e.preventDefault()}
            >
                <DialogHeader>
                    <DialogTitle>
                        {t("settings:reactivate.dialog.title")}
                    </DialogTitle>
                    <DialogDescription>
                        {t("settings:reactivate.dialog.description")}
                    </DialogDescription>
                </DialogHeader>

                <DialogFooter>
                    <Button
                        variant="outline"
                        disabled={isReactivating || isLoggingOut}
                        onClick={handleDecline}
                    >
                        {t("settings:reactivate.dialog.decline")}
                        {isLoggingOut && <IconLoader className="ml-1" />}
                    </Button>
                    <Button
                        disabled={isReactivating || isLoggingOut}
                        onClick={handleReactivate}
                    >
                        {t("settings:reactivate.dialog.confirm")}
                        {isReactivating && <IconLoader className="ml-1" />}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
