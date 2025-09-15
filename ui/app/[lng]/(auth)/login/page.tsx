"use client"

import Link from "next/link";
import IconGoogle from "@/components/icons/google";
import LogInForm from "@/components/forms/LoginForm";
import GLOBAL from "@/global";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect } from "react";
import { toast } from "react-toastify";
import ThemeSwitch from "@/components/utils/theme-switch";
import { useT } from "@/hooks/use-translation";

export default function LogInPage() {
    const { t } = useT("translation");
    const params = useSearchParams()
    const router = useRouter();

    useEffect(() => {
        const err = params.get("errorMessage")
        if (err) {
            toast(err, {
                type: "error"
            })
            return
        }
        
        const redirectURL = params.get("redirectUrl")
        if (redirectURL === null) return
        if (redirectURL.length === 0) router.push("/")
        else router.push(redirectURL);
    }, [params, router]);

    return (
        <section className="flex items-center justify-center h-screen w-screen bg-background">
            <ThemeSwitch className="absolute right-8 top-4" />
            <div className="flex flex-col justify-center rounded-xl p-12 w-full max-w-[600px]">
                <h1 className="text-lg font-bold">LET&apos;S LIVE</h1>
                <h1 className="text-2xl font-bold mb-1">{t("auth.login_title")}</h1>
                <p className="text-md">{t("auth.login_subtitle")}</p>
                <div className="flex gap-2 mb-2 mt-4">
                    <div className="w-full">
                        <Link
                            href={GLOBAL.API_URL + "/auth/google"}
                            className="h-12 flex-1 flex flex-row items-center justify-center gap-4 border border-border py-2 rounded-lg bg-white text-black hover:bg-[#ebebeb]"
                        >
                            <IconGoogle /> Google
                        </Link>
                        <p className="text-xs italic text-destructive mt-1">{t("auth.google_cookie_warning")}</p>
                    </div>
                </div>
                <div className="flex items-center justify-center w-full mt-2 mb-4">
                    <hr className="bg-border h-[2px] flex-1" />
                    <p className="text-center mx-4 text-foreground">{t("common.or")}</p>
                    <hr className="bg-border h-[2px] flex-1" />
                </div>
                <LogInForm />
                <p className="text-end text-sm opacity-80 mt-4">
                    {t("auth.no_account")}
                    <Link
                        href="/signup"
                        className="ml-2 text-blue-400 font-bold hover:text-blue-600"
                    >
                        {t("auth.signup")}
                    </Link>
                </p>
            </div>
        </section>
    );
}
