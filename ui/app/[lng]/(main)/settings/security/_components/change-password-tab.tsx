"use client";

import { useState } from "react";
import { toast } from "react-toastify";
import { ChangePassword } from "@/lib/api/auth";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import FormErrorText from "@/components/forms/FormErrorText";
import { Button } from "@/components/ui/button";
import IconLoader from "@/components/icons/loader";
import useT from "@/hooks/use-translation";
import { changePasswordSchema } from "@/lib/validations/changePassword";

export default function ChangePasswordTab() {
    const { t } = useT(["settings", "api-response"]);
    const [currentPassword, setCurrentPassword] = useState("");
    const [newPassword, setNewPassword] = useState("");
    const [confirmPassword, setConfirmPassword] = useState("");

    const [errors, setErrors] = useState({
        currentPassword: "",
        newPassword: "",
        confirmPassword: "",
    });
    const [isUpdatingPassword, setIsUpdatingPassword] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!validate()) return;

        setIsUpdatingPassword(true);
        const res = await ChangePassword({
            oldPassword: currentPassword,
            newPassword: newPassword,
        });

        setIsUpdatingPassword(false);

        if (!res.success) {
            setErrors((prev) => ({
                ...prev,
                confirmPassword: t(`api-response:${res.key}`),
            }));
            return;
        }

        setCurrentPassword("");
        setNewPassword("");
        setConfirmPassword("");
        setErrors({
            currentPassword: "",
            newPassword: "",
            confirmPassword: "",
        });

        toast(t("settings:security.security.password.updated_success"), {
            type: "success",
        });
    };

    const validate = () => {
        const result = changePasswordSchema(t).safeParse({
            currentPassword,
            newPassword,
            confirmPassword,
        });
        const newErrors: typeof errors = {
            currentPassword: "",
            newPassword: "",
            confirmPassword: "",
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

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
                <Label htmlFor="current-password">
                    {t(
                        "settings:security.security.password.form.current_label",
                    )}
                </Label>
                <Input
                    id="current-password"
                    type="password"
                    value={currentPassword}
                    onChange={(e) => setCurrentPassword(e.target.value)}
                    placeholder={t(
                        "settings:security.security.password.form.current_placeholder",
                    )}
                />
                <FormErrorText textError={errors.currentPassword} />
            </div>
            <div className="space-y-2">
                <Label htmlFor="new-password">
                    {t("settings:security.security.password.form.new_label")}
                </Label>
                <Input
                    id="new-password"
                    type="password"
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    placeholder={t(
                        "settings:security.security.password.form.new_placeholder",
                    )}
                />
                <FormErrorText textError={errors.newPassword} />
            </div>
            <div className="space-y-2">
                <Label htmlFor="confirm-password">
                    {t(
                        "settings:security.security.password.form.confirm_label",
                    )}
                </Label>
                <Input
                    id="confirm-password"
                    type="password"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    placeholder={t(
                        "settings:security.security.password.form.confirm_placeholder",
                    )}
                />
                <FormErrorText textError={errors.confirmPassword} />
            </div>
            <Button
                disabled={isUpdatingPassword}
                className="float-right disabled:bg-gray-300"
                type="submit"
            >
                {isUpdatingPassword && <IconLoader className="animate-spin" />}
                {t("settings:security.security.password.form.submit")}
            </Button>
        </form>
    );
}
