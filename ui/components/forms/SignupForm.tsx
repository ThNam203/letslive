"use client";

import { Loader } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "react-toastify";
import { RequestToSendVerification, SignUp, VerifyOTP } from "../../lib/api/auth";
import IconEmail from "../icons/email";
import FormErrorText from "./FormErrorText";
import IconUserOutline from "../icons/user";
import IconPasswordOutline from "../icons/password";
import IconEye from "../icons/eye";
import IconEyeOff from "../icons/eye-off";
import Turnstile, { useTurnstile } from "react-turnstile";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "../ui/dialog";
import { InputOTP, InputOTPGroup, InputOTPSlot } from "../ui/input-otp";
import { Button } from "../ui/button";
import { ResendOtpButton } from "./ResendButton";

export default function SignUpForm() {
    const [email, setEmail] = useState("");
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [hidingPassword, setHidingPassword] = useState(true);
    const [confirmPassword, setConfirmPassword] = useState("");
    const [turnstileToken, setTurnstileToken] = useState("");
    const [hidingConfirmPassword, setHidingConfirmPassword] = useState(true);
    const [isLoading, setIsLoading] = useState(false);
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

    const validate = () => {
        const newErrors = {
            email: "",
            password: "",
            confirmPassword: "",
            username: "",
            turnstile: "",
        };

        if (!email) {
            newErrors.email = "Email is required";
        } else if (!/\S+@\S+\.\S+/.test(email)) {
            newErrors.email = "Email is invalid";
        }

        if (!username) {
            newErrors.username = "Username is required";
        } else if (username.length < 6) {
            newErrors.username = "Username must be >= 6 characters";
        } else if (username.length > 20) {
            newErrors.username = "Username must be <= 20 characters";
        }

        if (!password) {
            newErrors.password = "Password is required";
        } else if (password.length < 8) {
            newErrors.password = "Password must be at least 8 characters";
        }

        if (!confirmPassword) {
            newErrors.confirmPassword = "Please confirm your password";
        } else if (confirmPassword !== password) {
            newErrors.confirmPassword = "Passwords do not match";
        }

        if (!turnstileToken) {
            newErrors.turnstile = "Please complete the CAPTCHA."
        }

        setErrors(newErrors);

        return (
            !newErrors.email &&
            !newErrors.password &&
            !newErrors.confirmPassword &&
            !newErrors.turnstile
        );
    };


    const handleSignUp = async () => {
        if (!validate()) {
            return;
        }

        if (otpValue.length !== 6) {
            setOtpError("Please enter a 6-digit OTP.");
            return;
        }

        setIsLoading(true);
        setIsOtpSubmitting(true);
        setOtpError("");

        const { fetchError } = await SignUp({
            email,
            username,
            password,
            turnstileToken,
            otpCode: otpValue,
        });

        if (fetchError) {
            setTurnstileToken("");
            turnstile.reset();
            setOtpValue("");
            toast.error(fetchError.message);
        } else {
            toast.success("Account created successfully");
            setIsOtpDialogOpen(false);
            router.replace("/")
            router.refresh()
        }

        setIsOtpDialogOpen(false);
        setIsOtpSubmitting(false);
        setIsLoading(false);
    };

    const handleBeginEmailVerification = async () => {
        setIsLoading(true);

        if (validate()) {
            const { fetchError } = await RequestToSendVerification(email, turnstileToken);
            if (fetchError) {
                turnstile.reset();
                setTurnstileToken("");
                toast.error(fetchError.message);
            } else {
                toast.success("Verification email sent, please check your inbox.");
                setIsOtpDialogOpen(true);
                setOtpValue("");
                setOtpError("");
            }
        }
        setIsLoading(false);
    };

    return (
        <div className="max-w">
            <form onSubmit={(e) => {
                e.preventDefault();
                e.stopPropagation();
                handleBeginEmailVerification();
            }}>
                <div className="flex px-4 gap-4 items-center rounded-md border border-border">
                    <label htmlFor="email">
                        <IconEmail className="opacity-40 scale-125" />
                    </label>
                    <input
                        id="email"
                        aria-label="Email"
                        className="h-12 focus:outline-none bg-background flex-1"
                        placeholder="Email"
                        type="email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                    />
                </div>
                <FormErrorText textError={errors.email} />
                <div className="flex px-4 gap-4 items-center rounded-md border border-border mt-4">
                    <label htmlFor="username">
                        <IconUserOutline className="opacity-40 scale-125" />
                    </label>
                    <input
                        id="username"
                        aria-label="Username"
                        className="h-12 focus:outline-none bg-background flex-1"
                        placeholder="Username"
                        type="text"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                    />
                </div>
                <FormErrorText textError={errors.username} />
                <div className="flex px-4 gap-4 items-center rounded-md border border-border mt-4">
                    <label htmlFor="password">
                        <IconPasswordOutline className="opacity-40 scale-125" />
                    </label>
                    <input
                        id="password"
                        aria-label="Password"
                        className="h-12 focus:outline-none bg-background flex-1"
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

                <div className="flex px-4 gap-4 items-center rounded-md border border-border mt-4">
                    <label htmlFor="confirm-password">
                        <IconPasswordOutline className="opacity-40 scale-125" />
                    </label>
                    <input
                        id="confirm-password"
                        aria-label="Confirm Password"
                        className="h-12 focus:outline-none bg-background flex-1"
                        placeholder="Confirm Password"
                        type={hidingConfirmPassword ? "password" : "text"}
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                    />
                    {!hidingConfirmPassword ? (
                        <IconEye
                            className="scale-150 opacity-50"
                            onClick={() => setHidingConfirmPassword(true)}
                        />
                    ) : (
                        <IconEyeOff
                            className="scale-150 opacity-50"
                            onClick={() => setHidingConfirmPassword(false)}
                        />
                    )}
                </div>
                <FormErrorText textError={errors.confirmPassword} />
                <Turnstile
                    sitekey={process.env.NEXT_PUBLIC_CLOUDFLARE_TURNSTILE_SITE_KEY!}
                    onSuccess={(token) => {
                        setTurnstileToken(token);
                        setErrors(prev => ({
                            ...prev,
                            turnstile: "",
                        }))
                    }}

                    onError={(err) => {
                        setTurnstileToken("");
                        setErrors(prev => ({
                            ...prev,
                            turnstile: err ?? ""
                        }))
                    }}
                    className="mt-4 my-2 float-right"
                />
                <FormErrorText textError={errors.turnstile} />
                <button
                    type="submit"
                    disabled={isLoading}
                    className="w-full rounded-md flex justify-center items-center bg-blue-400 hover:bg-blue-500 text-white h-12 border-transparent border font-semibold"
                >
                    {isLoading && <Loader className="animate-spin ml-2" />}
                    SIGN UP
                </button>
            </form>

            <Dialog open={isOtpDialogOpen} onOpenChange={setIsOtpDialogOpen}>
                <DialogContent onInteractOutside={(e) => { e.preventDefault() }}>
                    <DialogHeader>
                        <DialogTitle>Enter Verification Code</DialogTitle>
                        <DialogDescription>
                            A 6-digit code has been sent to{" "}
                            <span className="font-medium">{email}</span>. Please
                            enter it below to verify your email address.
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
                            <InputOTPGroup className="w-full flex">
                                <InputOTPSlot index={0} className="flex-1 h-14"/>
                                <InputOTPSlot index={1} className="flex-1 h-14"/>
                                <InputOTPSlot index={2} className="flex-1 h-14"/>
                                <InputOTPSlot index={3} className="flex-1 h-14"/>
                                <InputOTPSlot index={4} className="flex-1 h-14"/>
                                <InputOTPSlot index={5} className="flex-1 h-14"/>
                            </InputOTPGroup>
                        </InputOTP>
                        {otpError && (
                            <p className="text-sm text-destructive mt-2">{otpError}</p>
                        )}
                    </div>
                    <DialogFooter>
                        <ResendOtpButton onResend={handleBeginEmailVerification} initialCountdown={30}/>
                        <Button
                            type="button"
                            onClick={handleSignUp}
                            disabled={isOtpSubmitting || otpValue.length !== 6}
                            className="w-full"
                        >
                            {isOtpSubmitting && (
                                <Loader className="animate-spin mr-2 h-4 w-4" />
                            )}
                            VERIFY OTP
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}
