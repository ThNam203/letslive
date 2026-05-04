"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import useUser from "@/hooks/user";
import IconWallet from "@/components/icons/wallet";
import useT from "@/hooks/use-translation";

export default function WalletIcon() {
    const params = useParams();
    const lng = (params?.lng as string) ?? "en";
    const user = useUser((state) => state.user);
    const { t } = useT("accessibility");

    if (!user) return null;

    return (
        <Link
            href={`/${lng}/wallet/overview`}
            className="hover:bg-muted relative cursor-pointer rounded-md p-1.5 transition-colors"
            aria-label={t("wallet_open")}
            title={t("wallet")}
        >
            <IconWallet className="size-5" />
        </Link>
    );
}
