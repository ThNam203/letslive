"use client";

import { useState, useEffect } from "react";
import { toast } from "react-toastify";
import { RequestToSendVerification } from "../../../../lib/api/auth";

const VerificationRequest = () => {
    const [isCooldown, setIsCooldown] = useState(false);
    const [cooldownTime, setCooldownTime] = useState(60);

    useEffect(() => {
        if (isCooldown) {
            const timer = setInterval(() => {
                setCooldownTime((prev) => {
                    if (prev === 1) {
                        clearInterval(timer);
                        setIsCooldown(false);
                        return 60;
                    }
                    return prev - 1;
                });
            }, 1000);
            return () => clearInterval(timer);
        }
    }, [isCooldown]);

    const handleClick = async () => {
        if (!isCooldown) {
            setIsCooldown(true);

            const { fetchError } = await RequestToSendVerification();
            if (fetchError) {
                toast(fetchError.message, { type: "error" });
            }
        }
    };

    return (
        <p className="text-gray-500 text-sm">
            {isCooldown ? (
                <span className="text-gray-400">
                    Please wait {cooldownTime}s before retrying.
                </span>
            ) : (
                <span
                    className="underline text-purple-600 cursor-pointer hover:font-semibold"
                    onClick={handleClick}
                >
                    Click here
                </span>
            )}{" "}
            to send verification to your email.
        </p>
    );
};

export default VerificationRequest;
