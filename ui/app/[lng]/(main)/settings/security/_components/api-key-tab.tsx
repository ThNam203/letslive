"use client";

import { useState } from "react";
import { toast } from "@/components/utils/toast";
import useUser from "@/hooks/user";
import { RequestToGenerateNewAPIKey } from "@/lib/api/user";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import IconCopy from "@/components/icons/copy";
import IconRefresh from "@/components/icons/refresh";
import { cn } from "@/utils/cn";
import useT from "@/hooks/use-translation";

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
                    value={
                        user?.streamAPIKey
                            ? "•".repeat(user.streamAPIKey.length)
                            : ""
                    }
                    readOnly={true}
                    className="flex-grow border border-border text-right font-mono"
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
