"use client";

import Link from "next/link";
import IconGoogle from "@/components/icons/google";
import LogInForm from "@/components/forms/LoginForm";
import AccountDisabledDialog from "@/components/forms/AccountDisabledDialog";
import GLOBAL from "@/global";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";

export default function LogInPage() {
    const { t } = useT(["auth", "common"]);
    const searchParams = useSearchParams();
    const router = useRouter();
    const user = useUser((userState) => userState.user);
    const [reactivationToken, setReactivationToken] = useState<string | null>(
        null,
    );

    useEffect(() => {
        const err = searchParams.get("errorMessage");
        if (err) {
            toast(err, {
                type: "error",
            });
        }
    }, [searchParams, router]);

    useEffect(() => {
        const disabled = searchParams.get("accountDisabled");
        const token = searchParams.get("reactivationToken");
        if (disabled === "true" && token) {
            setReactivationToken(token);
        }
    }, [searchParams]);

    useEffect(() => {
        if (!user) return;
        const redirectUrl = searchParams.get("redirectUrl");
        if (
            redirectUrl &&
            redirectUrl.startsWith("/") &&
            !redirectUrl.startsWith("//")
        ) {
            router.push(redirectUrl);
            return;
        }
        router.push("/");
    }, [user, searchParams, router]);

    return (
        <>
            <h1 className="mb-1 text-2xl font-bold">{t("login_title")}</h1>
            <p className="text-md">{t("login_subtitle")}</p>
            <div className="mt-4 mb-2 flex gap-2">
                <div className="w-full">
                    <Link
                        href={GLOBAL.API_URL + "/auth/google"}
                        className="border-border flex h-12 flex-1 flex-row items-center justify-center gap-4 rounded-lg border bg-white py-2 text-black hover:bg-[#ebebeb]"
                    >
                        <IconGoogle /> Google
                    </Link>
                </div>
            </div>
            <div className="mt-2 mb-4 flex w-full items-center justify-center">
                <hr className="bg-border h-[2px] flex-1" />
                <p className="text-foreground mx-4 text-center">
                    {t("common:or")}
                </p>
                <hr className="bg-border h-[2px] flex-1" />
            </div>
            <LogInForm onAccountDisabled={setReactivationToken} />
            <AccountDisabledDialog
                reactivationToken={reactivationToken}
                onClose={() => setReactivationToken(null)}
            />
            <p className="mt-4 text-end text-sm opacity-80">
                {t("no_account")}
                <Link
                    href="/signup"
                    className="ml-2 font-bold text-blue-400 hover:text-blue-600"
                >
                    {t("signup")}
                </Link>
            </p>
        </>
    );
}
