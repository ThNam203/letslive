"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useWallet from "@/hooks/wallet";
import useUser from "@/hooks/user";
import { GetMyWallet, GetTransactions } from "@/lib/api/wallet";
import { CurrencyCode, Transaction } from "@/types/wallet";
import BalanceCard from "../_components/balance-card";
import TransactionRow from "../_components/transaction-row";
import Link from "next/link";
import IconLoader from "@/components/icons/loader";

export default function WalletOverviewPage() {
    const { t } = useT(["wallet", "api-response", "fetch-error"]);
    const user = useUser((s) => s.user);
    const { wallet, setWallet } = useWallet();
    const [recentTxns, setRecentTxns] = useState<Transaction[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    const fetchData = useCallback(async () => {
        if (!user) return;
        setIsLoading(true);
        try {
            const [walletRes, txnRes] = await Promise.all([
                GetMyWallet(),
                GetTransactions(0, 5),
            ]);

            if (walletRes.success && walletRes.data) {
                setWallet(walletRes.data);
            } else if (!walletRes.success) {
                toast.error(t(`api-response:${walletRes.key}`), {
                    toastId: walletRes.requestId,
                });
            }

            if (txnRes.success && txnRes.data) {
                setRecentTxns(txnRes.data);
            }
        } catch (_) {
            toast.error(t("fetch-error:client_fetch_error"));
        } finally {
            setIsLoading(false);
        }
    }, [user, setWallet, t]);

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    if (isLoading) {
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
