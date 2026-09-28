"use client";

import Link from "next/link";
import useUser from "@/hooks/user";
import IconWallet from "@/components/icons/wallet";
import useT from "@/hooks/use-translation";

export default function WalletIcon() {
    const user = useUser((state) => state.user);
    const { t } = useT("accessibility");

    return (
        <Link
            href={user ? "/wallet/overview" : "/login"}
            className="hover:bg-muted relative cursor-pointer rounded-md p-1.5 transition-colors"
            aria-label={t("wallet_open")}
            title={t("wallet")}
        >
            <IconWallet className="size-5" />
        </Link>
    );
}
