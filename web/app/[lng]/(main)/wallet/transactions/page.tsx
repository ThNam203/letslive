"use client";

import { useState } from "react";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { TransactionType } from "@/types/wallet";
import TransactionList from "../_components/transaction-list";
import { cn } from "@/utils/cn";
import { useTransactionsInfinite } from "@/hooks/queries/use-transactions";

const FILTER_OPTIONS: (TransactionType | "all")[] = [
    "all",
    TransactionType.DONATE,
    TransactionType.PURCHASE,
    TransactionType.REWARD,
    TransactionType.REFUND,
    TransactionType.TRADE,
    TransactionType.FEE,
];

export default function TransactionsPage() {
    const { t } = useT(["wallet", "api-response", "fetch-error"]);
    const user = useUser((s) => s.user);
    const [filter, setFilter] = useState<TransactionType | "all">("all");

    const { data, isLoading, isFetchingNextPage, hasNextPage, fetchNextPage } =
        useTransactionsInfinite(!!user);
    const transactions = data?.pages.flat() ?? [];

    const filteredTxns =
        filter === "all"
            ? transactions
            : transactions.filter((txn) => txn.type === filter);

    return (
        <section>
            <h2 className="text-foreground mb-1 text-xl font-semibold">
                {t("wallet:transactions.title")}
            </h2>
            <p className="text-muted-foreground mb-4 text-sm">
                {t("wallet:transactions.description")}
            </p>

            <div className="mb-4 flex flex-wrap gap-2">
                {FILTER_OPTIONS.map((opt) => {
                    const label =
                        opt === "all"
                            ? t("wallet:transactions.filter_all")
                            : t(`wallet:transactions.type.${opt}`);
                    return (
                        <button
                            key={opt}
                            onClick={() => setFilter(opt)}
                            className={cn(
                                "rounded-full px-3 py-1 text-xs font-medium transition-colors",
                                filter === opt
                                    ? "bg-primary text-primary-foreground"
                                    : "border-border text-muted-foreground hover:bg-background border",
                            )}
                        >
                            {label}
                        </button>
                    );
                })}
            </div>

            <div className="border-border rounded-lg border">
                <TransactionList
                    transactions={filteredTxns}
                    isLoading={isLoading || isFetchingNextPage}
                    hasMore={!!hasNextPage}
                    emptyMessage={t("wallet:transactions.no_transactions")}
                    emptyDescription={t(
                        "wallet:transactions.no_transactions_description",
                    )}
                    loadMoreLabel={t("wallet:transactions.load_more")}
                    onLoadMore={() => fetchNextPage()}
                />
            </div>
        </section>
    );
}
