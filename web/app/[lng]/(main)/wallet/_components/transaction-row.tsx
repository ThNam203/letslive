"use client";

import { cn } from "@/utils/cn";
import {
    Transaction,
    TransactionStatus,
    TransactionType,
} from "@/types/wallet";
import useT from "@/hooks/use-translation";

interface Props {
    transaction: Transaction;
}

const typeIcons: Record<TransactionType, string> = {
    [TransactionType.REWARD]: "&#127942;",
    [TransactionType.PURCHASE]: "&#128722;",
    [TransactionType.TRADE]: "&#128257;",
    [TransactionType.DONATE]: "&#10084;",
    [TransactionType.REFUND]: "&#8634;",
    [TransactionType.FEE]: "&#128176;",
    [TransactionType.ADJUSTMENT]: "&#9874;",
};

export default function TransactionRow({ transaction }: Props) {
    const { t } = useT("wallet");

    const typeLabel = t(`wallet:transactions.type.${transaction.type}`);
    const statusLabel = t(`wallet:transactions.status.${transaction.status}`);

    const netAmount = computeNetAmount(transaction);
    const isPositive = netAmount >= 0;

    return (
        <div className="border-border flex items-center justify-between border-b px-4 py-3 last:border-b-0">
            <div className="flex items-center gap-3">
                <span
                    className="text-lg"
                    dangerouslySetInnerHTML={{
                        __html: typeIcons[transaction.type] ?? "&#128176;",
                    }}
                />
                <div>
                    <p className="text-foreground text-sm font-medium">
                        {typeLabel}
                    </p>
                    <p className="text-muted-foreground text-xs">
                        {new Date(transaction.createdAt).toLocaleDateString(
                            undefined,
                            {
                                year: "numeric",
                                month: "short",
                                day: "numeric",
                                hour: "2-digit",
                                minute: "2-digit",
                            },
                        )}
                    </p>
                </div>
            </div>
            <div className="text-right">
                <p
                    className={cn(
                        "text-sm font-semibold",
                        isPositive ? "text-green-500" : "text-red-500",
                    )}
                >
                    {isPositive ? "+" : ""}
                    {netAmount.toLocaleString(undefined, {
                        minimumFractionDigits: 0,
                        maximumFractionDigits: 2,
                    })}
                </p>
                <StatusBadge status={transaction.status} label={statusLabel} />
            </div>
        </div>
    );
}

function StatusBadge({
    status,
    label,
}: {
    status: TransactionStatus;
    label: string;
}) {
    const color: Record<TransactionStatus, string> = {
        [TransactionStatus.CREATED]: "bg-gray-100 text-gray-700",
        [TransactionStatus.PROCESSING]: "bg-blue-100 text-blue-700",
        [TransactionStatus.COMPLETED]: "bg-green-100 text-green-700",
        [TransactionStatus.FAILED]: "bg-red-100 text-red-700",
        [TransactionStatus.CANCELLED]: "bg-yellow-100 text-yellow-700",
    };

    return (
        <span
            className={cn(
                "mt-0.5 inline-block rounded-full px-2 py-0.5 text-[10px] font-medium",
                color[status],
            )}
        >
            {label}
        </span>
    );
}

function computeNetAmount(transaction: Transaction): number {
    if (!transaction.entries || transaction.entries.length === 0) return 0;
    return transaction.entries.reduce(
        (sum, e) => sum + parseFloat(e.amount),
        0,
    );
}
