"use client";
import Link from "next/link";
import IconGoogle from "@/components/icons/google";
import GLOBAL from "@/global";
import SignUpForm from "@/components/forms/SignupForm";
import useT from "@/hooks/use-translation";

export default function SignUpPage() {
    const { t } = useT(["auth", "common"]);

    return (
        <>
            <h1 className="mb-1 text-2xl font-bold">{t("signup_title")}</h1>
            <p className="text-md">{t("signup_subtitle")}</p>
            <div className="mb-2 mt-4 flex gap-2">
                <div className="w-full">
                    <Link
                        href={GLOBAL.API_URL + "/auth/google"}
                        className="flex h-12 flex-1 flex-row items-center justify-center gap-4 rounded-lg border border-border bg-white py-2 text-black hover:bg-[#ebebeb]"
                    >
                        <IconGoogle /> Google
                    </Link>
                    {process.env.NEXT_PUBLIC_ENVIRONMENT === "production" && (
                        <p className="mt-1 text-xs italic text-destructive">
                            {t("google_cookie_warning")}
                        </p>
                    )}
                </div>
            </div>
            <div className="mb-4 mt-2 flex w-full items-center justify-center">
                <hr className="h-[2px] flex-1 bg-border" />
                <p className="mx-4 text-center text-foreground">
                    {t("common:or")}
                </p>
                <hr className="h-[2px] flex-1 bg-border" />
            </div>
            <SignUpForm />
            <p className="mt-4 text-end text-sm opacity-80">
                {t("have_account")}
                <Link
                    href="/login"
                    className="ml-2 font-bold text-blue-400 hover:text-blue-600"
                >
                    {t("login")}
                </Link>
            </p>
        </>
    );
}
