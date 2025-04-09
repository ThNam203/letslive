import React, { useState, useEffect, useCallback } from 'react';
import { Button } from "@/components/ui/button"; // Adjust the import path based on your project structure

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
  resendText?: string;
  countdownText?: (seconds: number) => string;
  sendingText?: string;
}

// --- The Component ---
export const ResendOtpButton: React.FC<ResendOtpButtonProps> = ({
  onResend,
  initialCountdown = 60,
  className,
  resendText = "Resend OTP",
  countdownText = (seconds) => `Resend in ${seconds}s`,
  sendingText = "Sending...",
}) => {
  const [countdown, setCountdown] = useState<number>(0);
  const [isResending, setIsResending] = useState<boolean>(false);

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
      console.error("Failed to resend OTP:", error);
    } finally {
      setIsResending(false); // Ensure loading state is reset
    }
  }, [onResend, initialCountdown, isButtonDisabled]); // Dependencies for useCallback

  const getButtonText = () => {
    if (isResending) {
      return sendingText;
    }
    if (countdown > 0) {
      return countdownText(countdown);
    }
    return resendText;
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