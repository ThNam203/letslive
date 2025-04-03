"use client";

import { useState } from "react";
import { toast } from "react-toastify";
import { Loader } from "lucide-react";
import { ChangePassword } from "../../../../../lib/api/auth";
import { Label } from "../../../../../components/ui/label";
import { Input } from "../../../../../components/ui/input";
import FormErrorText from "../../../../../components/forms/FormErrorText";
import { Button } from "../../../../../components/ui/button";

export default function ChangePasswordTab() {
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

    toast("Password updated successfully", { type: "success" });
  };

  const validatePassword = () => {
    if (newPassword !== confirmPassword) {
      setConfirmPasswordError("Passwords do not match");
      return false;
    } else setConfirmPasswordError("");

    if (newPassword.length < 8 || currentPassword.length < 8) {
      setUpdatePasswordError("Password must be at least 8 characters");
      return false;
    }

    if (newPassword === currentPassword) {
      setUpdatePasswordError(
        "New password must be different from current password"
      );
      return false;
    }

    return true;
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="current-password">Current Password</Label>
        <Input
          id="current-password"
          type="password"
          value={currentPassword}
          onChange={(e) => setCurrentPassword(e.target.value)}
          placeholder="Enter your current password"
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="new-password">New Password</Label>
        <Input
          id="new-password"
          type="password"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
          placeholder="Enter your new password"
        />
      </div>
      <div className="space-y-2">
        <Label htmlFor="confirm-password">Confirm New Password</Label>
        <Input
          id="confirm-password"
          type="password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          placeholder="Confirm your new password"
        />
      </div>
      <FormErrorText textError={confirmPasswordError} />
      <FormErrorText textError={updatePasswordError} />
      <Button
        disabled={isUpdatingPassword}
        className="disabled:bg-gray-300 float-right"
        type="submit"
      >
        {isUpdatingPassword && <Loader className="animate-spin" />}
        Confirm
      </Button>
    </form>
  );
}
