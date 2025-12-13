import { z } from "zod";
import i18next from "@/lib/i18n/i18next";
import { PASSWORD_MIN_LENGTH, PASSWORD_MAX_LENGTH } from "@/constant/password";

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
                .min(PASSWORD_MIN_LENGTH, t("error:password_too_short", { minLength: PASSWORD_MIN_LENGTH }))
                .max(PASSWORD_MAX_LENGTH, t("error:password_too_long", { maxLength: PASSWORD_MAX_LENGTH }))
                .refine(
                    (val) => /[a-z]/.test(val),
                    { message: t("error:password_missing_lowercase") }
                )
                .refine(
                    (val) => /[A-Z]/.test(val),
                    { message: t("error:password_missing_uppercase") }
                )
                .refine(
                    (val) => /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(val),
                    { message: t("error:password_missing_special") }
                ),
            confirmPassword: z
                .string()
                .min(1, t("error:confirm_password_required"))
                .min(PASSWORD_MIN_LENGTH, t("error:password_too_short", { minLength: PASSWORD_MIN_LENGTH }))
                .max(PASSWORD_MAX_LENGTH, t("error:password_too_long", { maxLength: PASSWORD_MAX_LENGTH })),
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
