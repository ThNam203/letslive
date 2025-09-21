import { z } from "zod";
import i18next from "@/lib/i18n/i18next";

export const signUpSchema = function (t: typeof i18next.t) {
    return z
        .object({
            email: z
                .email(t("error:email_invalid"))
                .min(1, t("error:email_required")),
            username: z
                .string()
                .min(1, t("error:username_required"))
                .min(6, t("error:username_too_short"))
                .max(20, t("error:username_too_long")),
            password: z
                .string()
                .min(1, t("error:password_required"))
                .min(8, t("error:password_too_short")),
            confirmPassword: z
                .string()
                .min(1, t("error:confirm_password_required")),
            turnstile: z.string().min(1, t("error:turnstile_required")),
        })
        .superRefine((data, ctx) => {
            if (data.confirmPassword !== data.password) {
                ctx.addIssue({
                    code: "custom",
                    message: t("error:passwords_do_not_match"),
                    path: ["confirmPassword"],
                });
            }
        });
};

export type SignUpSchema = z.infer<ReturnType<typeof signUpSchema>>;
