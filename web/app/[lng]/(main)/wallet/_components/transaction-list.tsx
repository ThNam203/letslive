"use client";

import { Transaction } from "@/types/wallet";
import { Button } from "@/components/ui/button";
import IconLoader from "@/components/icons/loader";
import TransactionRow from "./transaction-row";

interface Props {
    transactions: Transaction[];
    isLoading: boolean;
    hasMore: boolean;
    emptyMessage: string;
    emptyDescription?: string;
    loadMoreLabel: string;
    onLoadMore: () => void;
}

export default function TransactionList({
    transactions,
    isLoading,
    hasMore,
    emptyMessage,
    emptyDescription,
    loadMoreLabel,
    onLoadMore,
}: Props) {
    if (!isLoading && transactions.length === 0) {
        return (
            <div className="text-muted-foreground flex flex-col items-center justify-center py-12">
                <p className="text-sm font-medium">{emptyMessage}</p>
                {emptyDescription && (
                    <p className="mt-1 text-xs">{emptyDescription}</p>
                )}
            </div>
        );
    }

    return (
        <div>
            {transactions.map((txn) => (
                <TransactionRow key={txn.id} transaction={txn} />
            ))}
            {isLoading && (
                <div className="flex justify-center py-4">
                    <IconLoader />
                </div>
            )}
            {!isLoading && hasMore && (
                <div className="flex justify-center py-4">
                    <Button variant="outline" onClick={onLoadMore}>
                        {loadMoreLabel}
                    </Button>
                </div>
            )}
        </div>
    );
}
