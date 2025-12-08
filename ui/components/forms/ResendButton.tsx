"use client";

import React, { useState, useEffect, useCallback } from "react";
import { Button } from "@/components/ui/button"; // Adjust the import path based on your project structure
import { toast } from "react-toastify";
import useT from "@/hooks/use-translation";

// --- Component Props Interface ---
interface ResendOtpButtonProps {
    /**
     * Function to call when the resend button is clicked.
     * Should return a promise that resolves when the OTP request is sent
     * (or immediately if no async operation is needed on the frontend side).
     */
    onResend: () => Promise<void> | void;
    /**
     * The initial duration for the countdown timer in seconds.
     * @default 60
     */
    initialCountdown?: number;
    className?: string;
}

// --- The Component ---
export const ResendOtpButton: React.FC<ResendOtpButtonProps> = ({
    onResend,
    initialCountdown = 60,
    className,
}) => {
    const [countdown, setCountdown] = useState<number>(initialCountdown);
    const [isResending, setIsResending] = useState<boolean>(false);
    const { t } = useT(["auth", "error"]);

    const isButtonDisabled = countdown > 0 || isResending;

    useEffect(() => {
        let intervalId: NodeJS.Timeout | undefined;

        if (countdown > 0) {
            intervalId = setInterval(() => {
                setCountdown((prevCountdown) => prevCountdown - 1);
            }, 1000);
        } else {
            clearInterval(intervalId);
        }

        return () => clearInterval(intervalId);
    }, [countdown]);

    const handleResendClick = useCallback(async () => {
        if (isButtonDisabled) return; // Prevent clicking if disabled

        setIsResending(true);
        try {
            // Call the provided onResend function
            await onResend();
            // Start the countdown *after* the onResend logic completes successfully
            setCountdown(initialCountdown);
        } catch (error) {
            toast(t("error:otp_send_fail"), {
                toastId: "resend-otp-failed",
                type: "error",
            });
        } finally {
            setIsResending(false); // Ensure loading state is reset
        }
    }, [onResend, initialCountdown, isButtonDisabled]); // Dependencies for useCallback

    const getButtonText = () => {
        if (isResending) {
            return t("sending_otp");
        }
        if (countdown > 0) {
            return t("otp_resend_count_down", { countdown });
        }
        return t("resend_otp");
    };

    return (
        <Button
            type="button"
            variant="link"
            onClick={handleResendClick}
            disabled={isButtonDisabled}
            className={className}
        >
            {getButtonText()}
        </Button>
    );
};
