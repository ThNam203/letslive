import Link from "next/link";
import IconUpload from "@/components/icons/upload";
import { myGetT } from "@/lib/i18n";

export default async function UploadIcon() {
    const { t } = await myGetT("settings");

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
