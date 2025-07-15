import Link from "next/link";
import IconGoogle from "../../../components/icons/google";
import GLOBAL from "../../../global";
import SignUpForm from "@/components/forms/SignupForm";
import ThemeSwitch from "@/components/utils/theme-switch";

export default function SignUpPage() {
    return (
        <section className="flex h-screen w-screen items-center justify-center bg-background">
            <ThemeSwitch className="absolute right-8 top-4" />
            <div className="flex w-full max-w-[600px] flex-col justify-center rounded-xl p-12">
                <h1 className="text-lg font-bold">LET&apos;S LIVE</h1>
                <h1 className="mb-1 text-2xl font-bold">
                    Welcome! Sign up for a new world?
                </h1>
                <p className="text-md">Choose a method below to begin</p>
                <div className="mb-2 mt-4 flex gap-2">
                    <div className="w-full">
                        <Link
                                href={GLOBAL.API_URL + "/auth/google"}
                                className="h-12 flex-1 flex flex-row items-center justify-center gap-4 border border-border py-2 rounded-lg bg-white text-black hover:bg-[#ebebeb]"
                        >
                            <IconGoogle /> Google
                        </Link>
                        {process.env.NEXT_PUBLIC_ENVIRONMENT ===
                            "production" && (
                            <p className="text-destructive mt-1 text-xs italic">
                                Because the backend and frontend has different
                                domains, please allows 3rd party cookies to use
                                google authentication. I will fix it later.
                            </p>
                        )}
                    </div>
                </div>
                <div className="mb-4 mt-2 flex w-full items-center justify-center">
                    <hr className="h-[2px] flex-1 bg-border" />
                    <p className="mx-4 text-center text-foreground">or</p>
                    <hr className="h-[2px] flex-1 bg-border" />
                </div>
                <SignUpForm />
                <p className="mt-4 text-end text-sm opacity-80">
                    Already have an account?
                    <Link
                        href="/login"
                        className="ml-2 font-bold text-blue-400 hover:text-blue-600"
                    >
                        Log in
                    </Link>
                </p>
            </div>
        </section>
    );
}
