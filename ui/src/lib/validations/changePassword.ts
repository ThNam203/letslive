import { z } from "zod";
import i18next from "@/src/lib/i18n/i18next";
import { PASSWORD_MIN_LENGTH } from "@/src/constant/password";

export const changePasswordSchema = function (t: typeof i18next.t) {
    return z
        .object({
            currentPassword: z.string().min(PASSWORD_MIN_LENGTH, t("error:password_too_short", { minLength: PASSWORD_MIN_LENGTH })),
            newPassword: z.string().min(PASSWORD_MIN_LENGTH, t("error:password_too_short", { minLength: PASSWORD_MIN_LENGTH })),
            confirmPassword: z.string().min(PASSWORD_MIN_LENGTH, t("error:password_too_short", { minLength: PASSWORD_MIN_LENGTH })),
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

export type ChangePasswordSchema = z.infer<
    ReturnType<typeof changePasswordSchema>
>;
