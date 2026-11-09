"use client";

import IconEye from "../icons/eye";
import IconEyeOff from "../icons/eye-off";
import useT from "@/hooks/use-translation";

export default function PasswordVisibilityToggle({
    inputId,
    isVisible,
    onToggle,
}: {
    inputId: string;
    isVisible: boolean;
    onToggle: () => void;
}) {
    const { t } = useT("auth");

    return (
        <button
            type="button"
            aria-controls={inputId}
            aria-label={isVisible ? t("hide_password") : t("show_password")}
            onClick={onToggle}
            className="rounded-sm opacity-50 hover:cursor-pointer hover:opacity-80"
        >
            {isVisible ? <IconEye /> : <IconEyeOff />}
        </button>
    );
}
