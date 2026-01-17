"use client";

import { useState } from "react";
import { toast } from "react-toastify";
import useUser from "@/src/hooks/user";
import { RequestToGenerateNewAPIKey } from "@/src/lib/api/user";
import { Label } from "@/src/components/ui/label";
import { Input } from "@/src/components/ui/input";
import { Button } from "@/src/components/ui/button";
import IconCopy from "@/src/components/icons/copy";
import IconRefresh from "@/src/components/icons/refresh";
import { cn } from "@/src/utils/cn";
import useT from "@/src/hooks/use-translation";

export default function ApiKeyTab() {
    const { t } = useT("settings");
    const user = useUser((state) => state.user);
    const updateUser = useUser((state) => state.updateUser);
    const [isGenerating, setIsGenerating] = useState(false);

    const generateNewApiKey = async () => {
        if (!user) return;

        setIsGenerating(true);
        await RequestToGenerateNewAPIKey()
            .then((res) => {
                if (res.success) {
                    updateUser({
                        ...user,
                        streamAPIKey: res.data!,
                    });
                } else {
                    toast(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                        type: "error",
                    });
                }
            })
            .finally(() => {
                setIsGenerating(false);
            });
    };

    const copyApiKey = () => {
        if (!user) return;
        navigator.clipboard.writeText(user?.streamAPIKey);
        toast.success(t("settings:security.api_key.copied"));
    };

    return (
        <div>
            <div className="mb-4 flex flex-row items-center justify-between">
                <Label className="min-w-48" htmlFor="api-key">
                    {t("settings:security.api_key.label")}
                </Label>
                <Input
                    id="api-key"
                    value={user?.streamAPIKey}
                    readOnly={true}
                    className="flex-grow border border-border text-right"
                />
            </div>
            <div className="flex gap-4">
                <div className="flex-grow" />
                <Button disabled={isGenerating} onClick={copyApiKey}>
                    <IconCopy className="h-4 w-4 text-primary-foreground" />
                </Button>

                <Button disabled={isGenerating} onClick={generateNewApiKey}>
                    <IconRefresh
                        className={cn(
                            "mr-2 h-4 w-4",
                            isGenerating ? "animate-spin" : "",
                        )}
                    />{" "}
                    {t("settings:security.api_key.generate")}
                </Button>
            </div>
        </div>
    );
}
