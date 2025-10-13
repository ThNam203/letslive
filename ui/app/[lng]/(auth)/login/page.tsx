"use client";

import Link from "next/link";
import IconGoogle from "@/components/icons/google";
import LogInForm from "@/components/forms/LoginForm";
import GLOBAL from "@/global";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect } from "react";
import { toast } from "react-toastify";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";

export default function LogInPage() {
    const { t } = useT(["auth", "common"]);
    const searchParams = useSearchParams();
    const router = useRouter();
    const user = useUser((userState) => userState.user);

    useEffect(() => {
        const err = searchParams.get("errorMessage");
        if (err) {
            toast(err, {
                type: "error",
            });
            return;
        }

        const redirectURL = searchParams.get("redirectUrl");
        if (redirectURL === null) return;
        else router.push(redirectURL);
    }, [searchParams, router]);

    useEffect(() => {
        if (user) router.push("/");
    }, [user]);

    return (
        <>
            <h1 className="mb-1 text-2xl font-bold">{t("login_title")}</h1>
            <p className="text-md">{t("login_subtitle")}</p>
            <div className="mb-2 mt-4 flex gap-2">
                <div className="w-full">
                    <Link
                        href={GLOBAL.API_URL + "/auth/google"}
                        className="flex h-12 flex-1 flex-row items-center justify-center gap-4 rounded-lg border border-border bg-white py-2 text-black hover:bg-[#ebebeb]"
                    >
                        <IconGoogle /> Google
                    </Link>
                    <p className="mt-1 text-xs italic text-destructive">
                        {t("google_cookie_warning")}
                    </p>
                </div>
            </div>
            <div className="mb-4 mt-2 flex w-full items-center justify-center">
                <hr className="h-[2px] flex-1 bg-border" />
                <p className="mx-4 text-center text-foreground">
                    {t("common:or")}
                </p>
                <hr className="h-[2px] flex-1 bg-border" />
            </div>
            <LogInForm />
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
