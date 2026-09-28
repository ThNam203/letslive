"use client";

import Link from "next/link";
import IconWallet from "@/components/icons/wallet";
import useT from "@/hooks/use-translation";

export default function WalletIcon() {
    const { t } = useT("accessibility");

    return (
        <Link
            href="/wallet/overview"
            className="hover:bg-muted relative cursor-pointer rounded-md p-1.5 transition-colors"
            aria-label={t("wallet_open")}
            title={t("wallet")}
        >
            <IconWallet className="size-5" />
        </Link>
    );
}
