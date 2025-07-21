"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "react-toastify";
import { LogIn } from "../../lib/api/auth";
import IconEmail from "../icons/email";
import FormErrorText from "./FormErrorText";
import IconPasswordOutline from "../icons/password";
import IconEye from "../icons/eye";
import IconEyeOff from "../icons/eye-off";
import Turnstile, { useTurnstile } from "react-turnstile";
import IconLoader from "../icons/loader";

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
            newErrors.turnstile = "Please complete the CAPTCHA.";
        }

        setErrors(newErrors);

        return !newErrors.email && !newErrors.password && !newErrors.turnstile;
    };

    const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        e.stopPropagation();
        setIsLoading(true);

        if (validate()) {
            const { fetchError } = await LogIn({
                email,
                password,
                turnstileToken,
            });
            if (fetchError) {
                turnstile.reset();
                setTurnstileToken("");
                toast.error(fetchError.message);
            } else {
                router.replace("/");
                router.refresh();
            }
        }

        setIsLoading(false);
    };

    return (
        <form onSubmit={handleSubmit}>
            <div className="flex items-center gap-4 rounded-md border border-border px-4">
                <label htmlFor="email">
                    <IconEmail className="scale-125 opacity-40" />
                </label>
                <input
                    id="email"
                    aria-label="Email"
                    className="h-12 flex-1 bg-background focus:outline-none focus:bg-background"
                    autoComplete="email"
                    placeholder="Email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                />
            </div>
            <FormErrorText textError={errors.email} />
            <div className="mt-4 flex items-center gap-4 rounded-md border border-border px-4">
                <label htmlFor="password">
                    <IconPasswordOutline className="scale-125 opacity-40" />
                </label>
                <input
                    id="password"
                    aria-label="Password"
                    className="h-12 flex-1 bg-background focus:outline-none"
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
                    setErrors((prev) => ({
                        ...prev,
                        turnstile: "",
                    }));
                }}
                onError={(err) => {
                    setTurnstileToken("");
                    setErrors((prev) => ({
                        ...prev,
                        turnstile: err ?? "",
                    }));
                }}
                className="float-right my-2 mt-4"
            />
            <FormErrorText textError={errors.turnstile} />
            <button
                type="submit"
                disabled={isLoading}
                className="flex h-12 w-full items-center justify-center rounded-md border border-transparent bg-blue-400 font-semibold text-white hover:bg-blue-500"
            >
                {isLoading && <IconLoader className="ml-2" />}
                LOG IN
            </button>
        </form>
    );
}
