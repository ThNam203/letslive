"use client";

import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { CurrencyCode } from "@/types/wallet";
import BalanceCard from "../_components/balance-card";
import TransactionRow from "../_components/transaction-row";
import Link from "next/link";
import IconLoader from "@/components/icons/loader";
import { useRecentTransactions, useWalletBalance } from "@/hooks/queries/use-wallet";

export default function WalletOverviewPage() {
    const { t } = useT(["wallet", "api-response", "fetch-error"]);
    const user = useUser((s) => s.user);
    const { data: wallet, isLoading: isLoadingWallet } = useWalletBalance(!!user);
    const { data: recentTxns = [], isLoading: isLoadingTxns } =
        useRecentTransactions(!!user);

    if (isLoadingWallet || isLoadingTxns) {
        return (
            <div className="flex justify-center py-20">
                <IconLoader />
            </div>
        );
    }

    return (
        <>
            <section>
                <h2 className="text-foreground mb-4 text-xl font-semibold">
                    {t("wallet:overview.title")}
                </h2>
                <div className="grid gap-4 sm:grid-cols-2">
                    <BalanceCard
                        currencyCode={CurrencyCode.SPARK}
                        balance={
                            wallet?.balances.find(
                                (b) => b.currencyCode === CurrencyCode.SPARK,
                            )?.balance ?? "0"
                        }
                    />
                    <BalanceCard
                        currencyCode={CurrencyCode.FLARE}
                        balance={
                            wallet?.balances.find(
                                (b) => b.currencyCode === CurrencyCode.FLARE,
                            )?.balance ?? "0"
                        }
                    />
                </div>
            </section>

            <section>
                <div className="mb-4 flex items-center justify-between">
                    <h2 className="text-foreground text-xl font-semibold">
                        {t("wallet:overview.recent_transactions")}
                    </h2>
                    <Link
                        href="/wallet/transactions"
                        className="text-primary text-sm hover:underline"
                    >
                        {t("wallet:overview.view_all")}
                    </Link>
                </div>
                <div className="border-border rounded-lg border">
                    {recentTxns.length === 0 ? (
                        <p className="text-muted-foreground py-8 text-center text-sm">
                            {t("wallet:overview.no_transactions")}
                        </p>
                    ) : (
                        recentTxns.map((txn) => (
                            <TransactionRow key={txn.id} transaction={txn} />
                        ))
                    )}
                </div>
            </section>
        </>
    );
}
