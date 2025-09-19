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

export default function ChangePasswordTab() {
  const { t } = useT("settings");
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");

  const [updatePasswordError, setUpdatePasswordError] = useState("");
  const [confirmPasswordError, setConfirmPasswordError] = useState("");
  const [isUpdatingPassword, setIsUpdatingPassword] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validatePassword()) return;

    setIsUpdatingPassword(true);
    const { fetchError } = await ChangePassword({
      oldPassword: currentPassword,
      newPassword: newPassword,
    });

    setIsUpdatingPassword(false);

    if (fetchError) {
      setUpdatePasswordError(fetchError.message);
      return;
    }

    setCurrentPassword("");
    setNewPassword("");
    setConfirmPassword("");
    setUpdatePasswordError("");

    toast(t("settings:security.security.password.updated_success"), { type: "success" });
  };

  const validatePassword = () => {
    if (newPassword !== confirmPassword) {
      setConfirmPasswordError(t("settings:security.security.password.error_mismatch"));
      return false;
    } else setConfirmPasswordError("");

    if (newPassword.length < 8 || currentPassword.length < 8) {
      setUpdatePasswordError(t("settings:security.security.password.error_min_length"));
      return false;
    }

    if (newPassword === currentPassword) {
      setUpdatePasswordError(
        t("settings:security.security.password.error_same")
      );
      return false;
    }

    return true;
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="current-password">{t("settings:security.security.password.form.current_label")}</Label>
        <Input
          id="current-password"
          type="password"
          value={currentPassword}
          onChange={(e) => setCurrentPassword(e.target.value)}
          placeholder={t("settings:security.security.password.form.current_placeholder")}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="new-password">{t("settings:security.security.password.form.new_label")}</Label>
        <Input
          id="new-password"
          type="password"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
          placeholder={t("settings:security.security.password.form.new_placeholder")}
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="confirm-password">{t("settings:security.security.password.form.confirm_label")}</Label>
        <Input
          id="confirm-password"
          type="password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          placeholder={t("settings:security.security.password.form.confirm_placeholder")}
        />
      </div>
      <FormErrorText textError={confirmPasswordError} />
      <FormErrorText textError={updatePasswordError} />
      <Button
        disabled={isUpdatingPassword}
        className="disabled:bg-gray-300 float-right"
        type="submit"
      >
        {isUpdatingPassword && <IconLoader className="animate-spin" />}
        {t("settings:security.security.password.form.submit")}
      </Button>
    </form>
  );
}

