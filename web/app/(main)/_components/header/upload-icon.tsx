"use client";

import Link from "next/link";
import useT from "@/hooks/use-translation";
import IconUpload from "@/components/icons/upload";

export default function UploadIcon() {
    const { t } = useT("settings");

    return (
        <Link
            href="/settings/upload"
            className="hover:bg-muted relative cursor-pointer rounded-md p-1.5 transition-colors"
            aria-label={t("navigation.upload")}
            title={t("navigation.upload")}
        >
            <IconUpload className="size-5" />
        </Link>
    );
}
