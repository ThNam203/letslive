import { z } from "zod";
import i18next from "@/lib/i18n/i18next";

export const changePasswordSchema = function(t: typeof i18next.t) {
    return z
        .object({
            currentPassword: z
                .string()
                .min(8, t("error:password_too_short")),
            newPassword: z
                .string()
                .min(8, t("error:password_too_short")),
            confirmPassword: z
                .string()
                .min(8, t("error:password_too_short")),
        })
        .superRefine((data, ctx) => {
            if (data.newPassword === data.currentPassword) {
                ctx.addIssue({
                    code: "custom",
                    message: t("error:new_password_must_be_different"),
                    path: ["newPassword"],
                });
            }
            if (data.newPassword !== data.confirmPassword) {
                ctx.addIssue({
                    code: "custom",
                    message: t("error:passwords_do_not_match"),
                    path: ["confirmPassword"],
                });
            }
        });
};

export type ChangePasswordSchema = z.infer<ReturnType<typeof changePasswordSchema>>;
