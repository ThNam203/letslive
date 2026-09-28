"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import useUser from "@/hooks/user";
import IconWallet from "@/components/icons/wallet";
import useT from "@/hooks/use-translation";
import { I18N_FALLBACK_LNG } from "@/lib/i18n/settings";

export default function WalletIcon() {
    const params = useParams();
    const lng = (params?.lng as string) ?? I18N_FALLBACK_LNG;
    const user = useUser((state) => state.user);
    const { t } = useT("accessibility");

    return (
        <Link
            href={user ? `/${lng}/wallet/overview` : `/${lng}/login`}
            className="hover:bg-muted relative cursor-pointer rounded-md p-1.5 transition-colors"
            aria-label={t("wallet_open")}
            title={t("wallet")}
        >
            <IconWallet className="size-5" />
        </Link>
    );
}
