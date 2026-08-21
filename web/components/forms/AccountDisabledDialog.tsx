"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import IconLoader from "@/components/icons/loader";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { Reactivate } from "@/lib/api/auth";
import { GetMeProfile } from "@/lib/api/user";

interface AccountDisabledDialogProps {
    reactivationToken: string | null;
    onClose: () => void;
}

export default function AccountDisabledDialog({
    reactivationToken,
    onClose,
}: AccountDisabledDialogProps) {
    const [isReactivating, setIsReactivating] = useState(false);
    const { t } = useT(["auth", "api-response", "fetch-error"]);
    const { setUser } = useUser();
    const router = useRouter();

    const handleReactivate = async () => {
        if (!reactivationToken) return;

        setIsReactivating(true);
        await Reactivate({ reactivationToken })
            .then((res) => {
                if (!res.success) {
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                    });
                    onClose();
                    return;
                }

                GetMeProfile().then((meRes) => {
                    if (meRes.success && meRes.data) {
                        setUser(meRes.data);
                        router.push("/");
                    }
                });
                onClose();
            })
            .catch((_) => {
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "client-fetch-error-id",
                    type: "error",
                });
            })
            .finally(() => setIsReactivating(false));
    };

    return (
        <Dialog open={reactivationToken !== null}>
            <DialogContent
                className="bg-background text-foreground"
                showCloseButton={false}
                onInteractOutside={(e) => e.preventDefault()}
                onEscapeKeyDown={(e) => e.preventDefault()}
            >
                <DialogHeader>
                    <DialogTitle>
                        {t("auth:account_disabled_dialog_title")}
                    </DialogTitle>
                    <DialogDescription>
                        {t("auth:account_disabled_dialog_description")}
                    </DialogDescription>
                </DialogHeader>

                <DialogFooter>
                    <Button
                        variant="outline"
                        disabled={isReactivating}
                        onClick={onClose}
                    >
                        {t("auth:account_disabled_dialog_decline")}
                    </Button>
                    <Button disabled={isReactivating} onClick={handleReactivate}>
                        {t("auth:account_disabled_dialog_confirm")}
                        {isReactivating && <IconLoader className="ml-1" />}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
