"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import useUser from "@/hooks/user";
import { useUpdateProfile } from "@/hooks/queries/use-profile-mutations";
import useT from "@/hooks/use-translation";
import IconLoader from "@/components/icons/loader";
import IconUserOutline from "@/components/icons/user";
import { InputWithIconLabel } from "@/components/ui/input-with-icon-label";
import {
    USERNAME_MIN_LENGTH,
    USERNAME_MAX_LENGTH,
} from "@/constant/field-limits";

export default function AccountSetupPage() {
    const { t } = useT([
        "auth",
        "common",
        "error",
        "api-response",
        "fetch-error",
    ]);
    const router = useRouter();
    const user = useUser((s) => s.user);
    const isLoading = useUser((s) => s.isLoading);
    const updateProfile = useUpdateProfile();
    const isSubmitting = updateProfile.isPending;

    const [username, setUsername] = useState("");
    const [error, setError] = useState("");

    useEffect(() => {
        if (isLoading) return;
        if (!user) {
            router.replace("/login");
            return;
        }
        if (user.username !== "") {
            router.replace("/");
            return;
        }
    }, [user, isLoading, router]);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const trimmed = username.trim();
        if (trimmed.length < USERNAME_MIN_LENGTH) {
            setError(t("error:username_too_short"));
            return;
        }
        setError("");
        updateProfile.mutate(
            { username: trimmed },
            { onSuccess: () => router.replace("/") },
        );
    };

    if (isLoading || !user || user.username !== "") return null;

    return (
        <>
            <h1 className="mb-1 text-2xl font-bold">
                {t("auth:account_setup_title")}
            </h1>
            <p className="text-md mb-6 opacity-70">
                {t("auth:account_setup_subtitle")}
            </p>
            <form onSubmit={handleSubmit} className="flex flex-col gap-4">
                <div>
                    <InputWithIconLabel
                        icon={
                            <IconUserOutline className="scale-125 opacity-40" />
                        }
                        id="username"
                        aria-label={t("common:username")}
                        className="h-12 flex-1 border-none bg-transparent shadow-none focus-visible:ring-0"
                        placeholder={t("common:username")}
                        type="text"
                        maxLength={USERNAME_MAX_LENGTH}
                        value={username}
                        onChange={(e) => {
                            setUsername(e.target.value);
                            setError("");
                        }}
                        emitErrorSignalOnLimit={true}
                        disabled={isSubmitting}
                    />
                    {error && (
                        <p className="text-destructive mt-1 text-sm">{error}</p>
                    )}
                </div>
                <Button
                    type="submit"
                    disabled={isSubmitting || username.trim().length === 0}
                    className="mt-2 h-12 w-full"
                >
                    {isSubmitting && <IconLoader className="mr-2" />}
                    {t("auth:account_setup_submit")}
                </Button>
            </form>
        </>
    );
}
