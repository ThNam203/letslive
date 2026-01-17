"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "react-toastify";
import { RequestToSendVerification, SignUp } from "@/lib/api/auth";
import IconEmail from "@/components/icons/email";
import FormErrorText from "@/components/forms/FormErrorText";
import IconUserOutline from "@/components/icons/user";
import IconPasswordOutline from "@/components/icons/password";
import IconEye from "@/components/icons/eye";
import IconEyeOff from "@/components/icons/eye-off";
import Turnstile, { useTurnstile } from "react-turnstile";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp";
import { Button } from "@/components/ui/button";
import { ResendOtpButton } from "@/components/forms/ResendButton";
import IconLoader from "@/components/icons/loader";
import useT from "@/hooks/use-translation";
import { signUpSchema } from "@/lib/validations/signUp";
import { GetMeProfile } from "@/lib/api/user";
import useUser from "@/hooks/user";

export default function SignUpForm() {
    const [email, setEmail] = useState("");
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [hidingPassword, setHidingPassword] = useState(true);
    const [confirmPassword, setConfirmPassword] = useState("");
    const [turnstileToken, setTurnstileToken] = useState("");
    const [hidingConfirmPassword, setHidingConfirmPassword] = useState(true);
    const [isLoading, setIsLoading] = useState(false);
    const { setUser } = useUser();
    const router = useRouter();
    const [errors, setErrors] = useState({
        email: "",
        password: "",
        confirmPassword: "",
        username: "",
        turnstile: "",
    });
    const turnstile = useTurnstile();

    const [isOtpDialogOpen, setIsOtpDialogOpen] = useState(false);
    const [otpValue, setOtpValue] = useState("");
    const [isOtpSubmitting, setIsOtpSubmitting] = useState(false);
    const [otpError, setOtpError] = useState("");
    const { t, i18n } = useT([
        "auth",
        "error",
        "common",
        "api-response",
        "fetch-error",
    ]);

    const validate = () => {
        const parseResult = signUpSchema(t).safeParse({
            email,
            username,
            password,
            confirmPassword,
            turnstile: turnstileToken,
        });

        const newErrors: typeof errors = {
            email: "",
            username: "",
            password: "",
            confirmPassword: "",
            turnstile: "",
        };

        if (!parseResult.success) {
            for (const issue of parseResult.error.issues) {
                const pathKey = issue.path[0] as keyof typeof newErrors;
                if (pathKey && pathKey in newErrors) {
                    newErrors[pathKey] = issue.message;
                }
            }
        }

        setErrors(newErrors);
        return parseResult.success;
    };

    const handleSignUp = async () => {
        if (!validate()) {
            return;
        }

        if (otpValue.length !== 6) {
            setOtpError(t("otp_required"));
            return;
        }

        setIsLoading(true);
        setIsOtpSubmitting(true);
        setOtpError("");

        await SignUp({
            email,
            username,
            password,
            turnstileToken,
            otpCode: otpValue,
        })
            .then((res) => {
                if (!res.success) {
                    setTurnstileToken("");
                    turnstile.reset();
                    setOtpValue("");
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                    });
                } else {
                    toast.success(t("account_created_success"));
                    setIsOtpDialogOpen(false);

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
                setIsOtpSubmitting(false);
                setIsLoading(false);
            });
    };

    const handleBeginEmailVerification = async () => {
        if (!validate()) return;

        setIsLoading(true);
        await RequestToSendVerification(email, turnstileToken)
            .then((res) => {
                if (!res.success) {
                    turnstile.reset();
                    setTurnstileToken("");
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                    });
                } else {
                    toast.success(t(`api-response:${res.key}`));
                    setIsOtpDialogOpen(true);
                    setOtpValue("");
                    setOtpError("");
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
        <div className="max-w">
            <form
                onSubmit={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    handleBeginEmailVerification();
                }}
            >
                <div className="flex items-center gap-4 rounded-md border border-border px-4">
                    <label htmlFor="email">
                        <IconEmail className="scale-125 opacity-40" />
                    </label>
                    <input
                        id="email"
                        aria-label={t("email")}
                        className="h-12 flex-1 bg-background focus:outline-hidden"
                        placeholder={t("email")}
                        type="email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                    />
                </div>
                <FormErrorText textError={errors.email} />
                <div className="mt-4 flex items-center gap-4 rounded-md border border-border px-4">
                    <label htmlFor="username">
                        <IconUserOutline className="scale-125 opacity-40" />
                    </label>
                    <input
                        id="username"
                        aria-label={t("common:username")}
                        className="h-12 flex-1 bg-background focus:outline-hidden"
                        placeholder={t("common:username")}
                        type="text"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                    />
                </div>
                <FormErrorText textError={errors.username} />
                <div className="mt-4 flex items-center gap-4 rounded-md border border-border px-4">
                    <label htmlFor="password">
                        <IconPasswordOutline className="scale-125 opacity-40" />
                    </label>
                    <input
                        id="password"
                        aria-label={t("password")}
                        className="h-12 flex-1 bg-background focus:outline-hidden"
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

                <div className="mt-4 flex items-center gap-4 rounded-md border border-border px-4">
                    <label htmlFor="confirm-password">
                        <IconPasswordOutline className="scale-125 opacity-40" />
                    </label>
                    <input
                        id="confirm-password"
                        aria-label={t("confirm_password")}
                        className="h-12 flex-1 bg-background focus:outline-hidden"
                        placeholder={t("confirm_password")}
                        type={hidingConfirmPassword ? "password" : "text"}
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                    />
                    {!hidingConfirmPassword ? (
                        <IconEye
                            className="opacity-50 hover:cursor-pointer"
                            onClick={() => setHidingConfirmPassword(true)}
                        />
                    ) : (
                        <IconEyeOff
                            className="opacity-50 hover:cursor-pointer"
                            onClick={() => setHidingConfirmPassword(false)}
                        />
                    )}
                </div>
                <FormErrorText textError={errors.confirmPassword} />
                <div className="mt-4 flex h-[4.063rem] flex-col items-end">
                    <Turnstile
                        language={i18n.resolvedLanguage || i18n.language}
                        sitekey={
                            process.env
                                .NEXT_PUBLIC_CLOUDFLARE_TURNSTILE_SITE_KEY!
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
                    className="mt-4 flex h-12 w-full items-center justify-center rounded-md border border-transparent bg-blue-400 font-semibold uppercase text-white hover:bg-blue-500"
                >
                    {isLoading && <IconLoader className="ml-2" />}
                    {t("signup")}
                </button>
            </form>

            <Dialog open={isOtpDialogOpen} onOpenChange={setIsOtpDialogOpen}>
                <DialogContent
                    onInteractOutside={(e) => {
                        e.preventDefault();
                    }}
                >
                    <DialogHeader>
                        <DialogTitle>
                            {t("enter_verification_code")}
                        </DialogTitle>
                        <DialogDescription>
                            {t("otp_dialog_description_part_1")}{" "}
                            <span className="font-medium">{email}</span>
                            {t("otp_dialog_description_part_2")}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <InputOTP
                            id="otp-input"
                            maxLength={6}
                            value={otpValue}
                            onChange={(value) => setOtpValue(value)}
                            disabled={isOtpSubmitting}
                            onComplete={handleSignUp}
                            containerClassName="w-full"
                        >
                            <InputOTPGroup className="flex w-full">
                                <InputOTPSlot
                                    index={0}
                                    className="h-14 flex-1"
                                />
                                <InputOTPSlot
                                    index={1}
                                    className="h-14 flex-1"
                                />
                                <InputOTPSlot
                                    index={2}
                                    className="h-14 flex-1"
                                />
                                <InputOTPSlot
                                    index={3}
                                    className="h-14 flex-1"
                                />
                                <InputOTPSlot
                                    index={4}
                                    className="h-14 flex-1"
                                />
                                <InputOTPSlot
                                    index={5}
                                    className="h-14 flex-1"
                                />
                            </InputOTPGroup>
                        </InputOTP>
                        {otpError && (
                            <p className="mt-2 text-sm text-destructive">
                                {otpError}
                            </p>
                        )}
                    </div>
                    <DialogFooter>
                        <ResendOtpButton
                            onResend={handleBeginEmailVerification}
                            initialCountdown={60}
                        />
                        <Button
                            type="button"
                            onClick={handleSignUp}
                            disabled={isOtpSubmitting || otpValue.length !== 6}
                            className="w-full"
                        >
                            {isOtpSubmitting && (
                                <IconLoader className="mr-2 h-4 w-4" />
                            )}
                            {t("verify_otp")}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
