


import Link from "next/link";
import { IconGoogle } from "../../../components/icons/google";
import SignUpForm from "../../../components/forms/SignupForm";
import GLOBAL from "../../../global";

export default function SignUpPage() {
    return (
        <section className="flex items-center justify-center h-screen w-screen">
            <div className="flex flex-col justify-center rounded-xl p-12 bg-white w-full max-w-[600px]">
                <h1 className="text-lg font-bold">LET&apos;S LIVE</h1>
                <h1 className="text-2xl font-bold mb-1">
                    Welcome! Sign up for a new world?
                </h1>
                <p className="text-md">Choose a method below to begin</p>
                <div className="flex gap-2 mb-2 mt-4">
                    <div className="w-full">
                        <Link
                            className="flex-1 flex flex-row items-center justify-center gap-4 border-1 py-2 rounded-lg hover:bg-gray-200"
                            href={GLOBAL.API_URL + "/auth/google"}
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
                <SignUpForm />
                <p className="text-end text-sm opacity-80 mt-4">
                    Already have an account?
                    <Link
                        href="/login"
                        className="ml-2 text-blue-400 font-bold hover:text-blue-600"
                    >
                        Log in
                    </Link>
                </p>
            </div>
        </section>
    );
}
