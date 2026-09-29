import { ListSkeleton } from "@/components/skeletons/list-skeleton";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/utils/cn";

export function ConversationListSkeleton() {
    return <ListSkeleton rows={8} rowClassName="px-4 py-3" />;
}

export function ConversationHeaderSkeleton() {
    return (
        <div className="flex items-center gap-3 border-b p-4">
            <Skeleton className="h-10 w-10 rounded-full" />
            <div className="space-y-2">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-3 w-16" />
            </div>
        </div>
    );
}

const BUBBLES: { own: boolean; width: string }[] = [
    { own: false, width: "w-48" },
    { own: false, width: "w-64" },
    { own: true, width: "w-40" },
    { own: false, width: "w-56" },
    { own: true, width: "w-72" },
    { own: true, width: "w-32" },
];

export function MessageThreadSkeleton() {
    return (
        <div className="space-y-3 py-2" aria-busy="true">
            {BUBBLES.map((bubble, i) => (
                <div
                    key={i}
                    className={cn(
                        "flex",
                        bubble.own ? "justify-end" : "justify-start",
                    )}
                >
                    <Skeleton
                        className={cn(
                            "h-9 max-w-[75%] rounded-2xl",
                            bubble.width,
                        )}
                    />
                </div>
            ))}
        </div>
    );
}

/** The messages layout while the signed-in user is still loading. */
export function MessagesPageSkeleton({
    withThread = false,
}: {
    withThread?: boolean;
}) {
    return (
        <div className="flex h-full w-full">
            <div className="flex h-full w-full flex-col md:w-80 md:border-r">
                <div className="flex items-center gap-2 border-b p-4">
                    <Skeleton className="h-9 w-9" />
                    <Skeleton className="h-5 flex-1" />
                </div>
                <ConversationListSkeleton />
            </div>
            {withThread ? (
                <div className="hidden h-full flex-1 flex-col md:flex">
                    <ConversationHeaderSkeleton />
                    <div className="flex-1 px-4">
                        <MessageThreadSkeleton />
                    </div>
                </div>
            ) : (
                <div className="hidden flex-1 md:block" />
            )}
        </div>
    );
}
