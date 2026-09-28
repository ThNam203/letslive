"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import IconUpload from "@/components/icons/upload";
import { I18N_FALLBACK_LNG } from "@/lib/i18n/settings";

export default function UploadIcon() {
    const params = useParams();
    const lng = (params?.lng as string) ?? I18N_FALLBACK_LNG;
    const user = useUser((state) => state.user);
    const { t } = useT("settings");

    return (
        <Link
            href={user ? `/${lng}/settings/upload` : `/${lng}/login`}
            className="hover:bg-muted relative cursor-pointer rounded-md p-1.5 transition-colors"
            aria-label={t("navigation.upload")}
            title={t("navigation.upload")}
        >
            <IconUpload className="size-5" />
        </Link>
    );
}
