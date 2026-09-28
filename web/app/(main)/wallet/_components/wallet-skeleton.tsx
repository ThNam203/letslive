import { Skeleton } from "@/components/ui/skeleton";

export function BalanceCardsSkeleton() {
    return (
        <div className="grid gap-4 sm:grid-cols-2" aria-busy="true">
            {[0, 1].map((i) => (
                <div key={i} className="bg-muted space-y-3 rounded-xl p-6">
                    <Skeleton className="h-4 w-20" />
                    <Skeleton className="h-8 w-32" />
                    <Skeleton className="h-3 w-40" />
                </div>
            ))}
        </div>
    );
}

export function TransactionRowsSkeleton({ rows = 5 }: { rows?: number }) {
    return (
        <div aria-busy="true">
            {Array.from({ length: rows }, (_, i) => (
                <div
                    key={i}
                    className="border-border flex items-center justify-between border-b px-4 py-3 last:border-b-0"
                >
                    <div className="flex items-center gap-3">
                        <Skeleton className="h-6 w-6 rounded-full" />
                        <div className="space-y-2">
                            <Skeleton className="h-4 w-28" />
                            <Skeleton className="h-3 w-36" />
                        </div>
                    </div>
                    <Skeleton className="h-4 w-16" />
                </div>
            ))}
        </div>
    );
}
