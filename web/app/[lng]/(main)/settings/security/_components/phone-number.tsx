"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { UpdateProfile } from "@/lib/api/user";
import { useState } from "react";
import { toast } from "@/components/utils/toast";
import { parsePhoneNumberFromString } from "libphonenumber-js";
import { cn } from "@/utils/cn";
import { PHONE_NUMBER_MAX_LENGTH } from "@/constant/field-limits";

export default function PhoneNumber() {
    const { t } = useT("settings");
    const { user, updateUser } = useUser();
    const [isUpdating, setIsUpdating] = useState(false);
    const [phoneNumber, setPhoneNumber] = useState(user?.phoneNumber || "");
    const [isEditing, setIsEditing] = useState(false);

    const validatePhoneNumber = (value: string): boolean => {
        if (!value.trim()) return true; // ✅ allow empty
        const parsed = parsePhoneNumberFromString(value);
        return parsed ? parsed.isValid() : false;
    };

    const handleSave = async () => {
        // ✅ Only validate if not empty
        if (!validatePhoneNumber(phoneNumber)) {
            toast.error(t("settings:security.contact.phone_invalid"));
            return;
        }

        if (phoneNumber.trim() === user?.phoneNumber) {
            setIsEditing(false);
            return; // No changes made
        }

        try {
            setIsUpdating(true);
            const res = await UpdateProfile({
                phoneNumber: phoneNumber.trim() || "",
            });

            if (res.success) {
                updateUser({ ...user!, phoneNumber: phoneNumber.trim() || "" });
                setIsEditing(false);
                toast.success(t(`api-response:${res.key}`), {
                    toastId: res.requestId,
                });
            }
        } catch {
            toast.error(t("fetch-error:client_fetch_error"), {
                toastId: "client-fetch-error-id",
                type: "error",
            });
        } finally {
            setIsUpdating(false);
        }
    };

    return (
        <div className="border-border flex items-center justify-between border-t pt-4">
            <label
                className="min-w-48 text-sm font-medium"
                htmlFor="phone-number"
            >
                {t("settings:security.contact.phone")}
            </label>
            {!isEditing ? (
                <Button
                    onClick={() => {
                        setPhoneNumber(user?.phoneNumber || "");
                        setIsEditing(true);
                    }}
                    variant="none"
                    className={cn(
                        user?.phoneNumber
                            ? "text-medium text-foreground p-0 font-semibold italic"
                            : "text-primary hover:text-primary-hover p-0 text-sm",
                    )}
                >
                    {user?.phoneNumber
                        ? user.phoneNumber
                        : t("settings:security.contact.add_phone")}
                </Button>
            ) : (
                <Input
                    id="phone-number"
                    maxLength={PHONE_NUMBER_MAX_LENGTH}
                    value={phoneNumber ?? ""}
                    disabled={isUpdating}
                    onChange={(e) => {
                        setPhoneNumber(e.target.value);
                    }}
                    onKeyDown={(e) => {
                        if (e.key === "Enter") {
                            handleSave();
                        }
                    }}
                    placeholder="+1 234 567 8901"
                    autoFocus={true}
                    onBlur={() => {
                        setPhoneNumber(user?.phoneNumber || "");
                        setIsEditing(false);
                    }}
                    className="border-border flex-grow border text-right"
                />
            )}
        </div>
    );
}
