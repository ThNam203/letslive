import { z } from "zod";
import i18next from "@/lib/i18n/i18next";

export const loginSchema = function (t: typeof i18next.t) {
    return z.object({
        email: z
            .email(t("error:email_invalid"))
            .min(1, t("error:email_required")),
        password: z
            .string()
            .min(1, t("error:password_required"))
            .min(8, t("error:password_too_short")),
        turnstile: z.string().min(1, t("error:turnstile_required")),
    });
};

export type LoginSchema = z.infer<ReturnType<typeof loginSchema>>;
