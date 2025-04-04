"use client"

import Link from "next/link";
import { IconGoogle } from "../../../components/icons/google";
import LogInForm from "../../../components/forms/LoginForm";
import GLOBAL from "../../../global";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect } from "react";
import { toast } from "react-toastify";

export default function LogInPage() {
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
        <section className="flex items-center justify-center h-screen w-screen">
            <div className="flex flex-col justify-center rounded-xl p-12 bg-white w-full max-w-[600px]">
                <h1 className="text-lg font-bold">LET&apos;S LIVE</h1>
                <h1 className="text-2xl font-bold mb-1">Welcome back!</h1>
                <p className="text-md">Gain access to the world right now.</p>
                <div className="flex gap-2 mb-2 mt-4">
                    <div className="w-full">
                        <Link
                            href={GLOBAL.API_URL + "/auth/google"}
                            className="flex-1 flex flex-row items-center justify-center gap-4 border-1 py-2 rounded-lg hover:bg-gray-200"
                        >
                            <IconGoogle /> Google
                        </Link>
                        {process.env.NEXT_PUBLIC_ENVIRONMENT === "production" && <p className="text-xs italic text-red-500 mt-1">Because the backend and frontend has different domains, please allows 3rd party cookies to use google authentication. I will fix it later.</p>}
                    </div>
                </div>
                <div className="flex items-center justify-center w-full mt-2 mb-4">
                    <hr className="bg-gray-400 h-[2px] flex-1" />
                    <p className="text-center mx-4 text-gray-500">or</p>
                    <hr className="bg-gray-400 h-[2px] flex-1" />
                </div>
                <LogInForm />
                <p className="text-end text-sm opacity-80 mt-4">
                    Dont&#39;t have an account?
                    <Link
                        href="/signup"
                        className="ml-2 text-blue-400 font-bold hover:text-blue-600"
                    >
                        Sign up
                    </Link>
                </p>
            </div>
        </section>
    );
}
