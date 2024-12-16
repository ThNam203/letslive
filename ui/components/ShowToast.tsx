"use client";

import { useEffect } from "react";
import { toast } from "react-toastify";

const MISSING_MESSAGE = "Something went wrong!";

export default function ShowToast({
    id,
    err,
    success,
    info,
}: {
    id: string;
    err?: string;
    success?: string;
    info?: string;
}) {
    useEffect(() => {
        if (err !== undefined) {
            toast.error(formatMessage(err), {
                toastId: id || "error",
            });
            return;
        }

        if (success !== undefined) {
            toast.success(formatMessage(success), {
                toastId: id || "success",
            });
            return;
        }

        if (info !== undefined) {
            toast.info(formatMessage(info), {
                toastId: id || "info",
            });
            return;
        }

        toast.error("Something went wrong!", {});
    }, [err, success, info]); // Only trigger when `errMessage` changes
    return null;
}

function formatMessage(message?: string) {
    if (message === undefined || message.length === 0) return MISSING_MESSAGE;
    return message[0].toLowerCase() + message.slice(1);
}