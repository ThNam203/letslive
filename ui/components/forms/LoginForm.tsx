"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "react-toastify";
import { LogIn } from "../../lib/api/auth";
import { IconEmail } from "../icons/email";
import FormErrorText from "./FormErrorText";
import { IconPasswordOutline } from "../icons/password";
import { IconEye } from "../icons/eye";
import { IconEyeOff } from "../icons/eye-off";
import { Loader } from "lucide-react";
import Turnstile, { useTurnstile } from "react-turnstile";

export default function LogInForm() {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [hidingPassword, setHidingPassword] = useState(true);
    const [isLoading, setIsLoading] = useState(false);
    const router = useRouter();
    const [errors, setErrors] = useState({
        email: "",
        password: "",
        turnstile: "",
    });
    const [turnstileToken, setTurnstileToken] = useState("");
    const turnstile = useTurnstile();

    const validate = () => {
        const newErrors = { email: "", password: "", turnstile: "" };

        if (!email) {
            newErrors.email = "Email is required";
        } else if (!/\S+@\S+\.\S+/.test(email)) {
            newErrors.email = "Email is invalid";
        }

        if (!password) {
            newErrors.password = "Password is required";
        } else if (password.length < 8) {
            newErrors.password = "Password must be at least 8 characters";
        }
        
        if (!turnstileToken) {
            newErrors.turnstile = "Please complete the CAPTCHA."
        }

        setErrors(newErrors);

        return !newErrors.email && !newErrors.password && !newErrors.turnstile;
    };

    const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        e.stopPropagation();
        setIsLoading(true);

        if (validate()) {
            const { fetchError } = await LogIn({ email, password, turnstileToken });
            if (fetchError) {
                turnstile.reset();
                toast.error(fetchError.message);
            } else {
                router.replace("/")
                router.refresh()
            }
        }

        setIsLoading(false);
    };

    return (
        <form onSubmit={handleSubmit}>
            <div className="flex px-4 gap-4 items-center rounded-md border border-gray-200">
                <label htmlFor="email">
                    <IconEmail className="opacity-40 scale-125" />
                </label>
                <input
                    id="email"
                    aria-label="Email"
                    className="h-[50px] focus:outline-none flex-1"
                    placeholder="Email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                />
            </div>
            <FormErrorText textError={errors.email} />
            <div className="flex px-4 gap-4 items-center rounded-md border border-gray-200 mt-4">
                <label htmlFor="password">
                    <IconPasswordOutline className="opacity-40 scale-125" />
                </label>
                <input
                    id="password"
                    aria-label="Password"
                    className="h-[50px] focus:outline-none flex-1"
                    placeholder="Password"
                    type={hidingPassword ? "password" : "text"}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                />
                {!hidingPassword ? (
                    <IconEye
                        className="scale-150 opacity-50"
                        onClick={() => setHidingPassword(true)}
                    />
                ) : (
                    <IconEyeOff
                        className="scale-150 opacity-50"
                        onClick={() => setHidingPassword(false)}
                    />
                )}
            </div>
            <FormErrorText textError={errors.password} />
            <Turnstile
                sitekey={process.env.NEXT_PUBLIC_CLOUDFLARE_TURNSTILE_SITE_KEY!}
                onSuccess={(token) => {
                    setTurnstileToken(token);
                    setErrors(prev => ({
                        ...prev,
                        turnstile: "",
                    }))
                }}
                onError={(err) => setErrors(prev => ({
                    ...prev,
                    turnstile: err ?? ""
                }))}
                className="mt-4 my-2 float-right"
            />
            <FormErrorText textError={errors.turnstile} />
            <button
                type="submit"
                disabled={isLoading}
                className="w-full rounded-md flex justify-center items-center bg-blue-400 hover:bg-blue-500 text-white h-[50px] border-transparent border font-semibold"
            >
                {isLoading && <Loader className="animate-spin ml-2" />}
                LOG IN
            </button>
        </form>
    );
}
