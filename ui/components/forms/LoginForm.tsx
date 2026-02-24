"use client";

import { useParams, useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "@/components/utils/toast";
import { LogIn } from "@/lib/api/auth";
import IconEmail from "../icons/email";
import FormErrorText from "./FormErrorText";
import IconPasswordOutline from "../icons/password";
import IconEye from "../icons/eye";
import IconEyeOff from "../icons/eye-off";
import Turnstile, { useTurnstile } from "react-turnstile";
import IconLoader from "../icons/loader";
import useT from "@/hooks/use-translation";
import { loginSchema } from "@/lib/validations/login";
import { GetMeProfile } from "@/lib/api/user";
import useUser from "@/hooks/user";

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
    const { t, i18n } = useT(["auth", "error", "api-response", "fetch-error"]);
    const { setUser } = useUser();
    const validate = () => {
        const result = loginSchema(t).safeParse({
            email,
            password,
            turnstile: turnstileToken,
        });
        const newErrors: typeof errors = {
            email: "",
            password: "",
            turnstile: "",
        };
        if (!result.success) {
            for (const issue of result.error.issues) {
                const key = issue.path[0] as keyof typeof newErrors;
                if (key in newErrors) newErrors[key] = issue.message;
            }
        }
        setErrors(newErrors);
        return result.success;
    };

    const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        e.stopPropagation();

        if (!validate()) return;

        setIsLoading(true);
        await LogIn({
            email,
            password,
            turnstileToken,
        })
            .then((res) => {
                if (!res.success) {
                    turnstile.reset();
                    setTurnstileToken("");
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                    });
                } else {
                    GetMeProfile().then((res) => {
                        if (res.success && res.data) {
                            setUser(res.data);
                            router.push("/");
                        }
                    });
                }
            })
            .catch((_) => {
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "client-fetch-error-id",
                    type: "error",
                });
            })
            .finally(() => {
                setIsLoading(false);
            });
    };

    return (
        <form onSubmit={handleSubmit}>
            <div className="border-border flex items-center gap-4 rounded-md border px-4">
                <label htmlFor="email">
                    <IconEmail className="scale-125 opacity-40" />
                </label>
                <input
                    id="email"
                    aria-label="Email"
                    className="bg-background focus:bg-background h-12 flex-1 focus:outline-none"
                    autoComplete="email"
                    placeholder={t("email")}
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                />
            </div>
            <FormErrorText textError={errors.email} />
            <div className="border-border mt-4 flex items-center gap-4 rounded-md border px-4">
                <label htmlFor="password">
                    <IconPasswordOutline className="scale-125 opacity-40" />
                </label>
                <input
                    id="password"
                    aria-label="Password"
                    className="bg-background h-12 flex-1 focus:outline-none"
                    placeholder={t("password")}
                    type={hidingPassword ? "password" : "text"}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                />
                {!hidingPassword ? (
                    <IconEye
                        className="opacity-50 hover:cursor-pointer"
                        onClick={() => setHidingPassword(true)}
                    />
                ) : (
                    <IconEyeOff
                        className="opacity-50 hover:cursor-pointer"
                        onClick={() => setHidingPassword(false)}
                    />
                )}
            </div>
            <FormErrorText textError={errors.password} />
            <div className="mt-4 flex h-[4.063rem] flex-col items-end">
                <Turnstile
                    language={i18n.resolvedLanguage || i18n.language}
                    sitekey={
                        process.env.NEXT_PUBLIC_CLOUDFLARE_TURNSTILE_SITE_KEY!
                    }
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
                />
                <FormErrorText textError={errors.turnstile} />
            </div>
            <button
                type="submit"
                disabled={isLoading}
                className="mt-4 flex h-12 w-full items-center justify-center rounded-md border border-transparent bg-blue-400 font-semibold text-white uppercase hover:bg-blue-500"
            >
                {isLoading && <IconLoader className="ml-2" />}
                {t("login")}
            </button>
        </form>
    );
}
