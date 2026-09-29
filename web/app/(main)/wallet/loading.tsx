import {
    BalanceCardsSkeleton,
    TransactionRowsSkeleton,
} from "./_components/wallet-skeleton";

export default function WalletLoading() {
    return (
        <>
            <BalanceCardsSkeleton />
            <div className="border-border rounded-lg border">
                <TransactionRowsSkeleton />
            </div>
        </>
    );
}
