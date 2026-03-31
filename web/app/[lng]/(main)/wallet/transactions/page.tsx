"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { GetTransactions } from "@/lib/api/wallet";
import { Transaction, TransactionType } from "@/types/wallet";
import TransactionList from "../_components/transaction-list";
import { cn } from "@/utils/cn";

const PAGE_SIZE = 20;

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
    const [transactions, setTransactions] = useState<Transaction[]>([]);
    const [page, setPage] = useState(0);
    const [total, setTotal] = useState(0);
    const [isLoading, setIsLoading] = useState(true);
    const [filter, setFilter] = useState<TransactionType | "all">("all");

    const fetchTransactions = useCallback(
        async (pageNum: number) => {
            if (!user) return;
            setIsLoading(true);
            try {
                const res = await GetTransactions(pageNum, PAGE_SIZE);
                if (res.success && res.data) {
                    if (pageNum === 0) {
                        setTransactions(res.data);
                    } else {
                        setTransactions((prev) => [...prev, ...res.data!]);
                    }
                    if (res.meta) {
                        setTotal(res.meta.total ?? 0);
                    }
                } else if (!res.success) {
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                    });
                }
            } catch (_) {
                toast.error(t("fetch-error:client_fetch_error"));
            } finally {
                setIsLoading(false);
            }
        },
        [user, t],
    );

    useEffect(() => {
        fetchTransactions(0);
    }, [fetchTransactions]);

    const handleLoadMore = useCallback(() => {
        const nextPage = page + 1;
        setPage(nextPage);
        fetchTransactions(nextPage);
    }, [page, fetchTransactions]);

    const filteredTxns =
        filter === "all"
            ? transactions
            : transactions.filter((txn) => txn.type === filter);
    const hasMore = transactions.length < total;

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
                                    : "border-border text-muted-foreground border hover:bg-background",
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
                    isLoading={isLoading}
                    hasMore={hasMore}
                    emptyMessage={t("wallet:transactions.no_transactions")}
                    emptyDescription={t(
                        "wallet:transactions.no_transactions_description",
                    )}
                    loadMoreLabel={t("wallet:transactions.load_more")}
                    onLoadMore={handleLoadMore}
                />
            </div>
        </section>
    );
}
