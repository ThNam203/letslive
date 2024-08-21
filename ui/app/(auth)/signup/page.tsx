import OAuthLoginBtn from "@/components/buttons/OAuthLoginBtn";
import SignUpForm from "@/components/forms/SignupForm";
import { IconFacebook } from "@/components/icons/facebook";
import { IconGoogle } from "@/components/icons/google";

import GLOBAL from "@/global";

export default function SignUpPage() {
    return (
        <section className="flex items-center justify-center h-screen w-screen">
            <div className="flex flex-col justify-center rounded-xl p-12 bg-white">
                <h1 className="text-lg font-bold">LET&apos;S LIVE</h1>
                <h1 className="text-2xl font-bold mb-1">
                    Welcome! Sign up for a new world?
                </h1>
                <p className="text-md">Choose a method below to begin</p>
                <div className="flex gap-2 mb-2 mt-4">
                    <OAuthLoginBtn url={GLOBAL.API_URL + "/v1/auth/google"}>
                        <IconGoogle/> Google
                    </OAuthLoginBtn>
                    <OAuthLoginBtn url={GLOBAL.API_URL + "/v1/auth/google"}>
                        <IconFacebook/> Facebook
                    </OAuthLoginBtn>
                </div>
                <div className="flex items-center justify-center w-full mt-2 mb-4">
                    <hr className="bg-gray-400 h-[2px] flex-1" />
                    <p className="text-center mx-4 text-gray-500">or</p>
                    <hr className="bg-gray-400 h-[2px] flex-1" />
                </div>
                <SignUpForm />
                <p className="text-end text-sm opacity-80 mt-4">
                    Already have an account?
                    <a
                        href="/login"
                        className="ml-2 text-blue-400 font-bold hover:text-blue-600"
                    >
                        Log in
                    </a>
                </p>
            </div>
        </section>
    );
}
